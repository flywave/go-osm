package osm

import (
	"math/rand"
	"sort"
)

var colorkeys = []string{"#0030E5", "#0042E4", "#0053E4", "#0064E4", "#0075E4", "#0186E4", "#0198E3", "#01A8E3", "#01B9E3", "#01CAE3", "#02DBE3", "#02E2D9", "#02E2C8", "#02E2B7", "#02E2A6", "#03E295", "#03E184", "#03E174", "#03E163", "#03E152", "#04E142", "#04E031", "#04E021", "#04E010", "#09E004", "#19E005", "#2ADF05", "#3BDF05", "#4BDF05", "#5BDF05", "#6CDF06", "#7CDE06", "#8CDE06", "#9DDE06", "#ADDE06", "#BDDE07", "#CDDD07", "#DDDD07", "#DDCD07", "#DDBD07", "#DCAD08", "#DC9D08", "#DC8D08", "#DC7D08", "#DC6D08", "#DB5D09", "#DB4D09", "#DB3D09", "#DB2E09", "#DB1E09", "#DB0F0A"}
var sizecolorkeys = len(colorkeys)

func RandomColor() string {
	return colorkeys[rand.Intn(sizecolorkeys)]
}

func Reverse(s []int) []int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func Satisify2(ring1 []int, ring2 []int) bool {
	_, lastid1 := ring1[0], ring1[len(ring1)-1]
	firstid2, _ := ring2[0], ring2[len(ring2)-1]
	return firstid2 == lastid1
}

func Collision(ring1 []int, ring2 []int) ([]int, bool, bool) {
	firstid1, lastid1 := ring1[0], ring1[len(ring1)-1]
	firstid2, lastid2 := ring2[0], ring2[len(ring2)-1]
	total := []int{}
	boolval := false
	if firstid1 == firstid2 {
		total = append(ring1, Reverse(ring2)...)
		boolval = true
	} else if firstid1 == lastid2 {
		total = append(ring2, ring1...)
		boolval = true
	} else if lastid1 == lastid2 {
		total = append(ring1, Reverse(ring2)...)
		boolval = true
	} else if lastid1 == firstid2 {
		total = append(ring1, ring2...)
		boolval = true
	}
	if len(total) == 0 {
		return []int{}, false, false
	}

	return total, boolval, total[0] != total[len(total)-1]
}

func Satisfy(member []int) bool {
	return member[0] == member[len(member)-1]

}

func SortedMap(mymap map[int][]int) []int {
	newlist := make([]int, len(mymap))
	pos := 0
	for k := range mymap {
		newlist[pos] = k
		pos++
	}
	sort.Ints(newlist)
	return Reverse(newlist)
}
func cleanse(member []int) []int {
	if len(member) > 0 {
		if member[0] == member[len(member)-1] {
			return member[:len(member)-1]
		} else {
			return member
		}
	} else {
		return member
	}
}

func Connect(members [][]int) [][]int {

	membermap := map[int][]int{}
	totalmembers := [][]int{}
	for pos, member := range members {
		if Satisfy(member) {
			totalmembers = append(totalmembers, member)
		} else {
			membermap[pos] = member
		}
	}

	generation := 0
	for len(membermap) > 2 && generation < 100 {

		for _, k := range SortedMap(membermap) {
			member, boolval1 := membermap[k]
			boolval := true
			if boolval1 {
				lastpt := member[len(member)-1]
				for _, ktry := range SortedMap(membermap) {
					trymember, boolval2 := membermap[ktry]

					if boolval2 {
						if k != ktry && boolval == true {
							if lastpt == trymember[0] {

								if len(membermap) == 2 {

								} else {
									membermap[k] = append(member, trymember...)

									delete(membermap, ktry)

								}

								boolval = true
							}
						}
					}
				}
			}
		}
		generation += 1

	}

	generation = 0
	for len(membermap) > 2 && generation < 100 {

		for _, k := range SortedMap(membermap) {
			member, boolval1 := membermap[k]
			if boolval1 {
				boolval := false
				for _, ktry := range SortedMap(membermap) {
					trymember, boolval2 := membermap[ktry]
					if boolval2 {
						if k != ktry && boolval == false {
							twomember := len(member) == 2 && len(membermap) <= 4

							if len(membermap) == 2 {
								if member[len(member)-1] != trymember[0] {
									membermap[k] = append(member, Reverse(trymember)...)
								} else {
									membermap[k] = append(member, trymember...)

								}
								delete(membermap, ktry)
							}

							if member[0] == trymember[0] && !twomember {
								membermap[ktry] = Reverse(trymember)
								boolval = true
							} else if member[len(member)-1] == trymember[len(trymember)-1] && !twomember {
								membermap[ktry] = Reverse(trymember)
								boolval = true
							} else if member[0] == trymember[len(trymember)-1] {
								boolval = true
							} else if member[len(member)-1] == trymember[0] {
								membermap[k] = append(member, trymember...)
								boolval = true

								delete(membermap, ktry)
							} else {

							}
						}

					}
				}
			}
		}
		generation += 1

	}
	if len(membermap) == 2 {
		var member, trymember []int
		var pos int
		var k, ktry int
		for kk, v := range membermap {
			if pos == 0 {
				pos = 1
				member = v
				k = kk
			} else if pos == 1 {
				trymember = v
				ktry = kk
			}
		}
		if member[len(member)-1] != trymember[0] {
			membermap[k] = append(member, Reverse(trymember)...)
		} else {
			membermap[k] = append(member, trymember...)

		}
		delete(membermap, ktry)
	}
	pos := 0
	for _, v := range membermap {
		totalmembers = append(totalmembers, v)
		pos++
	}

	return totalmembers

}

func ConvertNodes(nodes []int, nodemap map[int][]float64) [][]float64 {
	ring := make([][]float64, len(nodes))
	for pos, node := range nodes {
		ring[pos] = nodemap[node]
	}
	return ring
}
