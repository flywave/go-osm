package osm

import (
	"fmt"
	"sort"
)

type SizeSorter []WayBlock

func (a SizeSorter) Len() int           { return len(a) }
func (a SizeSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SizeSorter) Less(i, j int) bool { return a[i].Size < a[j].Size }

type DifSorter []WayBlock

func (a DifSorter) Len() int           { return len(a) }
func (a DifSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DifSorter) Less(i, j int) bool { return a[i].Dif < a[j].Dif }

func SortKeysString(mymap map[int]string) []int {
	i := 0
	newlist := make([]int, len(mymap))
	for k := range mymap {
		newlist[i] = k
		i++
	}
	sort.Ints(newlist)
	return newlist
}

func get_uniques(totalmap map[int]map[int]string) (map[int]string, map[int]int) {
	newmap := map[int]string{}
	indmap := map[int]int{}
	for _, v := range totalmap {
		for k := range v {
			newmap[k] = ""
		}
	}

	for pos, k := range SortKeysString(newmap) {
		indmap[k] = pos
	}

	return newmap, indmap
}

type WayBlock struct {
	Position int
	Binary   []byte
	Nodes    []int
	Dif      int
	Size     int
}

func SingleDif(xs, ys []byte) int {
	total := 0
	for i := range xs {
		x, y := xs[i], ys[i]
		if x != y {
			total++
		}
	}
	return total
}

func SingleFill(xs, ys []byte) int {
	total := 0
	for i := range xs {
		x, y := xs[i], ys[i]
		if x == 1 && y == 0 {
			total++
		}
	}
	return total
}

func GetCoverage(xs, ys []byte) ([]byte, int) {
	total := 0
	newlist := make([]byte, len(xs))
	for pos := range xs {
		x, y := xs[pos], ys[pos]
		if x == 1 || y == 1 {
			newlist[pos] = 1
			total++
		} else {
			newlist[pos] = 0
		}
	}
	return newlist, total
}

func NewWayBlock(pos int, nodemap map[int]string, bytevals []byte, indmap map[int]int) WayBlock {
	tmp := make([]byte, len(bytevals))
	copy(tmp, bytevals)
	nodes := make([]int, len(nodemap))
	i := 0
	for k := range nodemap {
		pos := indmap[k]
		tmp[pos] = 1
		nodes[i] = k
		i++
	}
	return WayBlock{Position: pos, Size: len(nodemap), Nodes: nodes, Binary: tmp}
}

type Pathing struct {
	WayBlocks  []WayBlock
	Binary     []byte
	PathBlocks []WayBlock
	Uniques    map[int]string
	TotalMap   map[int]map[int]string
	Limit      int
}

func (pathing *Pathing) SortSize() {
	sort.Sort(sort.Reverse(SizeSorter(pathing.WayBlocks)))
}

func (pathing *Pathing) SortDif() {
	sort.Sort(sort.Reverse(DifSorter(pathing.WayBlocks)))
}

type OutputDif struct {
	Pos int
	Dif int
}

func (pathing *Pathing) GetDif(ys []byte, myfunc func(xs, ys []byte) int) {
	c := make(chan OutputDif)
	if len(pathing.WayBlocks) < 1000 {
		for pos, wayblock := range pathing.WayBlocks {
			go func(pos int, wayblock WayBlock, c chan OutputDif) {
				c <- OutputDif{Pos: pos, Dif: myfunc(wayblock.Binary, ys)}
			}(pos, wayblock, c)
		}
		for range pathing.WayBlocks {
			output := <-c
			pathing.WayBlocks[output.Pos].Dif = output.Dif
		}
	} else {
		bump := int(float64(len(pathing.WayBlocks)) * .9)
		for pos, wayblock := range pathing.WayBlocks[bump:] {
			go func(pos int, wayblock WayBlock, c chan OutputDif) {
				c <- OutputDif{Pos: pos, Dif: myfunc(wayblock.Binary, ys)}
			}(pos, wayblock, c)
		}
		for range pathing.WayBlocks[bump:] {
			output := <-c
			pathing.WayBlocks[output.Pos].Dif = output.Dif
		}
	}

	pathing.SortDif()
}

func (pathing *Pathing) GetMinDif() {
	row := pathing.PathBlocks[len(pathing.PathBlocks)-1]
	pathing.GetDif(row.Binary, SingleDif)
	addrow := pathing.WayBlocks[len(pathing.WayBlocks)-1]
	pathing.PathBlocks = append(pathing.PathBlocks, addrow)
	pathing.WayBlocks = pathing.WayBlocks[:len(pathing.WayBlocks)-1]
}

func (pathing *Pathing) GetTopDif() WayBlock {
	pathing.SortSize()
	wayblock1 := pathing.WayBlocks[0]
	return wayblock1
}

func (pathing *Pathing) AddRemove(wayblock WayBlock) {
	newlist := make([]WayBlock, len(pathing.WayBlocks)-1)
	increment := 0
	for _, i := range pathing.WayBlocks {
		if i.Position != wayblock.Position {
			newlist[increment] = i
			increment++
		}
	}
	pathing.WayBlocks = newlist
	pathing.PathBlocks = append(pathing.PathBlocks, wayblock)
}

func (pathing *Pathing) TotalDif() int {
	oldi := pathing.PathBlocks[0]
	totaldif := 0
	for _, i := range pathing.PathBlocks[1:] {
		totaldif += SingleDif(oldi.Binary, i.Binary)
		oldi = i
	}
	return totaldif
}

func (pathing *Pathing) Path() {
	wayblock := pathing.GetTopDif()
	pathing.AddRemove(wayblock)
	totalmapsize := len(pathing.TotalMap)
	fmt.Println()
	for len(pathing.WayBlocks) != 0 {
		pathing.GetMinDif()
		fmt.Printf("\rCreating Shortest Path [%d/%d]", len(pathing.PathBlocks), totalmapsize)
	}
	fmt.Println()
}

type ReadWay struct {
	Ways  []int
	Nodes map[int]string
}

func (pathing *Pathing) CreateReads() []ReadWay {
	nodemap := map[int]string{}
	tempexisting := map[int]string{}
	tempnew := map[int]string{}
	tempways := []int{}
	totalreads := []ReadWay{}
	for pos, way := range pathing.PathBlocks {
		for _, node := range way.Nodes {
			_, boolval := nodemap[node]
			if boolval {
				tempexisting[node] = ""
			} else {
				tempnew[node] = ""
			}
		}
		tempways = append(tempways, way.Position)

		if len(tempexisting)+len(tempnew) > pathing.Limit || pos == len(pathing.PathBlocks)-1 {
			newnodemap := map[int]string{}
			for k := range tempexisting {
				newnodemap[k] = ""
			}
			for k := range tempnew {
				newnodemap[k] = ""
			}

			if len(newnodemap) > pathing.Limit || pos == len(pathing.PathBlocks)-1 {
				totalreads = append(totalreads, ReadWay{Ways: tempways, Nodes: newnodemap})
				nodemap = newnodemap
				tempways = []int{}
				tempexisting, tempnew = map[int]string{}, map[int]string{}

			}
		}
	}
	return totalreads
}

func NewPathing(totalmap map[int]map[int]string, limit int) Pathing {
	uniques, indmap := get_uniques(totalmap)
	bytevals := make([]byte, len(uniques))
	count := 0
	wayblocks := make([]WayBlock, len(totalmap))
	for k, v := range totalmap {
		wayblocks[count] = NewWayBlock(k, v, bytevals, indmap)
		count++
	}
	return Pathing{Binary: bytevals, WayBlocks: wayblocks, Uniques: uniques, Limit: limit, TotalMap: totalmap}
}

func MakeReads(totalmap map[int]map[int]string, limit int) []ReadWay {
	pathing := NewPathing(totalmap, limit)
	pathing.Path()
	return pathing.CreateReads()
}
