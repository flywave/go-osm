package osm

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"math"
	"os"

	"github.com/flywave/go-geom"

	"github.com/flywave/go-osm/osmpbf"
)

type Poly [][]float64

func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func RoundPt(pt []float64) []float64 {
	return []float64{Round(pt[0], .5, 6), Round(pt[1], .5, 6)}
}

func (c Poly) Pip(p []float64) bool {
	intersections := 0
	for i := range c {
		curr := c[i]
		ii := i + 1
		if ii == len(c) {
			ii = 0
		}
		next := c[ii]

		bottom, top := curr, next
		if bottom[1] >= top[1] {
			bottom, top = top, bottom
		}
		if p[1] <= bottom[1] || p[1] >= top[1] {
			continue
		}

		if p[0] >= math.Max(curr[0], next[0]) ||
			next[1] == curr[1] {
			continue
		}

		xint := (p[1]-curr[1])*(next[0]-curr[0])/(next[1]-curr[1]) + curr[0]
		if curr[0] != next[0] && p[0] > xint {
			continue
		}

		intersections++
	}
	return intersections%2 != 0
}

func (poly Poly) Within(inner Poly) bool {
	boolval := true
	for _, pt := range inner {
		if !poly.Pip(pt) {
			boolval = false
			return boolval
		}
	}
	return boolval
}

func (d *Decoder) CreateTestCaseBlock(key int, idmap map[int]string) {
	primblock := d.Relations[key]
	relmap := d.ReadRelationsLazy(primblock)
	totalbool := false
	for _, relation := range d.ReadBlock(*primblock).Primitivegroup[0].Relations {
		_, boolval := idmap[int(relation.Id)]
		if boolval {
			totalbool = true
		}
	}

	if totalbool {
		totalmap := map[int]string{}
		for k := range relmap {
			val, boolval := d.Ways[k]
			if boolval {
				tempmap := d.ReadWaysLazy(val, d.IdMap)
				for k := range tempmap {
					totalmap[k] = ""
				}
			}
		}

		stringval := make([]*LazyPrimitiveBlock, len(totalmap))
		i := 0
		for k := range totalmap {
			stringval[i] = d.DenseNodes[k]
			i++
		}
		d.SyncWaysNodeMapMultiple(stringval, d.IdMap)

		pb := d.ReadBlock(*primblock)
		relations := pb.Primitivegroup[0].Relations
		waymap := map[int][]int{}
		for _, way := range relations {
			refs := way.Memids
			oldref := refs[0]
			pos := 1
			newrefs := make([]int, len(refs))
			newrefs[0] = int(refs[0])
			for _, ref := range refs[1:] {
				ref = ref + oldref
				newrefs[pos] = int(ref)
				pos++
				oldref = ref
			}

			for _, i := range newrefs {
				waymap[d.WayIdMap.GetBlock(i)] = append(waymap[d.WayIdMap.GetBlock(i)], i)
			}
		}

		totalwaynodemap := map[int][]int{}
		for k, v := range waymap {
			val, boolval := d.Ways[k]
			if boolval {
				tempwaynodemap := d.ReadWaysLazyList(val, v)
				for k, v := range tempwaynodemap {
					totalwaynodemap[k] = v
				}
			}
		}

		for ipos, way := range relations {
			_, mybool := idmap[int(way.Id)]
			if mybool {
				refs := way.Memids
				oldref := refs[0]
				pos := 1
				newrefs := make([]int, len(refs))
				newrefs[0] = int(refs[0])
				for _, ref := range refs[1:] {
					ref = ref + oldref
					newrefs[pos] = int(ref)
					pos++
					oldref = ref
				}

				totalidmap := map[int]string{}
				for _, i := range newrefs {

					vals, boolval := totalwaynodemap[i]
					if boolval {
						for _, nodeid := range vals {
							totalidmap[d.IdMap.GetBlock(nodeid)] = ""
						}
					}

				}

				add_nodes := make([]int, len(totalidmap))
				newpos := 0
				for k := range totalidmap {
					add_nodes[newpos] = k
					newpos++
				}

				if len(add_nodes) > 0 {
					d.AddUpdates(add_nodes)
				}
				fmt.Println(len(add_nodes), ipos)

				st := pb.GetStringtable().GetS()

				roles := make([]string, len(way.RolesSid))
				for pos, ri := range way.RolesSid {
					roles[pos] = string(st[int(ri)])
				}

				inners := [][]int{}
				outers2 := [][]int{}
				for i := range newrefs {
					role, wayid := roles[i], newrefs[i]
					nodes, boolval := totalwaynodemap[wayid]

					if boolval {
						if role == "inner" {
							inners = append(inners, nodes)
						} else if role == "outer" {
							outers2 = append(outers2, nodes)
						}
					}
				}

				total := [][][]int{outers2, inners}
				var network bytes.Buffer

				enc := gob.NewEncoder(&network)
				err := enc.Encode(total)
				if err != nil {
					fmt.Println(err)
				}

				outfilenodes := fmt.Sprintf("test_cases/%d.gob", int(way.Id))
				ioutil.WriteFile(outfilenodes, network.Bytes(), 0677)

				nodemap := map[int][]float64{}
				for _, nodes := range inners {
					for _, node := range nodes {
						nodemap[node] = d.GetNode(node)
					}
				}
				for _, nodes := range outers2 {
					for _, node := range nodes {
						nodemap[node] = d.GetNode(node)
					}
				}

				var network2 bytes.Buffer
				enc = gob.NewEncoder(&network2)
				err = enc.Encode(nodemap)
				if err != nil {
					fmt.Println(err)
				}

				outfilenodemap := fmt.Sprintf("test_cases/%d_nodemap.gob", int(way.Id))
				ioutil.WriteFile(outfilenodemap, network2.Bytes(), 0677)
			}
		}
	}
}

func (d *Decoder) ProcessRelationBlock(key int, blockcount int) {
	primblock := d.Relations[key]

	pb := d.ReadBlock(*primblock)
	relations := pb.Primitivegroup[0].Relations
	waymap := map[int][]int{}
	st := pb.GetStringtable().GetS()

	for _, way := range relations {
		refs := way.Memids
		oldref := refs[0]
		pos := 1
		newrefs := make([]int, len(refs))
		newrefs[0] = int(refs[0])
		for _, ref := range refs[1:] {
			ref = ref + oldref
			newrefs[pos] = int(ref)
			pos++
			oldref = ref
		}

		mymap := map[string]interface{}{}
		for i := range way.Keys {
			keypos, valpos := way.Keys[i], way.Vals[i]
			mymap[string(st[keypos])] = st[valpos]
		}
		if mymap["type"] == "multipolygon" {
			for _, i := range newrefs {
				wayblock := d.WayIdMap.GetBlock(i)
				waymap[wayblock] = append(waymap[wayblock], i)
			}
		} else {
		}
	}

	totalwaynodemap := map[int][]int{}
	for k, v := range waymap {
		val, boolval := d.Ways[k]
		if boolval {
			tempwaynodemap := d.ReadWaysLazyList(val, v)
			for k, v := range tempwaynodemap {

				totalwaynodemap[k] = v
			}
		}
	}

	totalidmap := map[int]string{}

	temp_relations := []*osmpbf.Relation{}

	sizerels := len(relations)
	for ipos, way := range relations {
		refs := way.Memids
		oldref := refs[0]
		pos := 1
		newrefs := make([]int, len(refs))
		newrefs[0] = int(refs[0])
		for _, ref := range refs[1:] {
			ref = ref + oldref
			newrefs[pos] = int(ref)
			pos++
			oldref = ref
		}

		for _, i := range newrefs {
			vals, boolval := totalwaynodemap[i]
			if boolval {
				for _, nodeid := range vals {
					totalidmap[d.IdMap.GetBlock(nodeid)] = ""
				}
			}
		}

		temp_relations = append(temp_relations, &way)
		if len(totalidmap) > d.Limit || ipos == sizerels-1 {

			add_nodes := make([]int, len(totalidmap))
			newpos := 0
			for k := range totalidmap {
				add_nodes[newpos] = k
				newpos++
			}

			if len(add_nodes) > 0 {
				d.AddUpdates(add_nodes)
			}

			fmt.Printf("\r[%d/%d] Relation Blocks [%d/%d] Relations Read in this block. Memory Throughput: %dmb", blockcount+1, len(d.Relations), ipos, len(relations), d.TotalMemory/1000000)
			for _, way := range temp_relations {
				refs := way.Memids
				oldref := refs[0]
				pos := 1
				newrefs := make([]int, len(refs))
				newrefs[0] = int(refs[0])
				for _, ref := range refs[1:] {
					ref = ref + oldref
					newrefs[pos] = int(ref)
					pos++
					oldref = ref
				}

				roles := make([]string, len(way.RolesSid))
				for pos, ri := range way.RolesSid {
					roles[pos] = string(st[int(ri)])
				}

				mymap := map[string]interface{}{}
				for i := range way.Keys {
					keypos, valpos := way.Keys[i], way.Vals[i]
					mymap[string(st[keypos])] = st[valpos]
				}

				mymap["osm_id"] = int(way.Id)

				if mymap["type"] == "multipolygon" {
					inners := [][]int{}
					outers2 := [][]int{}
					for i := range newrefs {
						d.RelationMap[newrefs[i]] = ""

						role, wayid := roles[i], newrefs[i]
						nodes, boolval := totalwaynodemap[wayid]

						if boolval {
							if role == "inner" {
								inners = append(inners, nodes)
							} else if role == "outer" {
								outers2 = append(outers2, nodes)
							}
						}
					}

					inners = Connect(inners)
					outers2 = Connect(outers2)
					innermap := map[int][][]float64{}
					outers := [][][]float64{}
					for pos, inner := range inners {
						ring := make([][]float64, len(inner))
						for pos, node := range inner {
							ring[pos] = RoundPt(d.GetNode(node))
						}
						innermap[pos] = ring
					}

					for _, outer := range outers2 {
						ring := make([][]float64, len(outer))
						for pos, node := range outer {
							ring[pos] = RoundPt(d.GetNode(node))
						}
						outers = append(outers, ring)
					}

					polygons := [][][][]float64{}
					for _, outer := range outers {
						newpolygon := [][][]float64{outer}
						for id, inner := range innermap {
							boolval := Poly(outer).Within(Poly(inner))
							if boolval {
								newpolygon = append(newpolygon, inner)
								delete(innermap, id)
							}
						}
						polygons = append(polygons, newpolygon)
					}
					if len(polygons) > 0 {
						if len(polygons) == 1 {
							featpolygon := geom.NewPolygonFeature(polygons[0])
							featpolygon.Properties = mymap
							if d.WriteBool {
								d.Writer.WriteFeature(featpolygon)
							}
						} else {
							featpolygon := geom.NewMultiPolygonFeature(polygons...)
							featpolygon.Properties = mymap
							if d.WriteBool {
								d.Writer.WriteFeature(featpolygon)
							}
						}
					}
				}
			}
			temp_relations = []*osmpbf.Relation{}
			totalidmap = map[int]string{}
		}
	}
}

func (d *Decoder) ProcessRelations() {
	fmt.Println()
	relationlist := SortKeys(d.Relations)

	for i, key := range relationlist {
		d.ProcessRelationBlock(key, i)
	}
	fmt.Println()
}

func (d *Decoder) CreateTestCases(ids []int) {
	os.MkdirAll("test_cases", os.ModePerm)

	idmap := map[int]string{}
	for _, id := range ids {
		idmap[id] = ""
	}

	relationlist := SortKeys(d.Relations)

	sizerelation := len(relationlist)
	for i, key := range relationlist {
		d.CreateTestCaseBlock(key, idmap)
		fmt.Printf("\r[%d/%d] Processing Relations", i, sizerelation)
	}
}
