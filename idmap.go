package osm

import (
	"fmt"
	"math"
	"strconv"
)

var Power9 = math.Pow(10.0, 9.0)
var Power8 = math.Pow(10.0, 8.0)
var Power7 = math.Pow(10.0, 7.0)
var Power6 = math.Pow(10.0, 6.0)

type IdMap struct {
	IdMap map[int]map[int]map[int]map[int][]*LazyPrimitiveBlock
}

func (idmap *IdMap) AddKey(k9, k8, k7, k6 int, block *LazyPrimitiveBlock) {
	_, boolval := idmap.IdMap[k9]
	if !boolval {
		idmap.IdMap[k9] = map[int]map[int]map[int][]*LazyPrimitiveBlock{}
	}
	_, boolval = idmap.IdMap[k9][k8]
	if !boolval {
		idmap.IdMap[k9][k8] = map[int]map[int][]*LazyPrimitiveBlock{}
	}
	_, boolval = idmap.IdMap[k9][k8][k7]
	if !boolval {
		idmap.IdMap[k9][k8][k7] = map[int][]*LazyPrimitiveBlock{}
	}
	idmap.IdMap[k9][k8][k7][k6] = append(idmap.IdMap[k9][k8][k7][k6], block)
}

func GetVal(k9, k8, k7, k6 int) int {
	stringval := fmt.Sprintf("%d%d%d%d", k9, k8, k7, k6)
	if string(stringval[0]) == "0" && string(stringval[1]) == "0" && string(stringval[2]) == "0" {
		stringval = stringval[3:]
	} else if string(stringval[0]) == "0" && string(stringval[1]) == "0" {
		stringval = stringval[2:]

	} else if string(stringval[0]) == "0" {
		stringval = stringval[1:]
	}
	val, _ := strconv.ParseInt(stringval, 0, 64)

	return int(val)
}

func Convstr(valstr string) int {
	val, _ := strconv.ParseInt(valstr, 0, 64)
	return int(val)
}

func GetStrKeys(val int) (int, int, int, int) {
	valstr := strconv.Itoa(val)
	if len(valstr) < 4 {
		needed := 4 - len(valstr)
		strneeded := ""
		for i := 0; i < needed; i++ {
			strneeded += "0"
		}
		valstr = strneeded + valstr
	}

	k9str, k8str, k7str, k6str := string(valstr[0]), string(valstr[1]), string(valstr[2]), string(valstr[3])
	k9, k8, k7, k6 := Convstr(k9str), Convstr(k8str), Convstr(k7str), Convstr(k6str)
	return k9, k8, k7, k6
}

func (idmap *IdMap) AddBlock(block *LazyPrimitiveBlock) {
	k19, k18, k17, k16 := GetKeys(block.IdRange[0])
	k29, k28, k27, k26 := GetKeys(block.IdRange[1])
	if k19 == k29 && k18 == k28 && k17 == k27 && k16 == k26 {
		idmap.AddKey(k19, k18, k17, k16, block)
	} else {
		idmap.AddKey(k19, k18, k17, k16, block)
		idmap.AddKey(k29, k28, k27, k26, block)

		keyint1 := GetVal(k19, k18, k17, k16)
		keyint2 := GetVal(k29, k28, k27, k26)

		for i := keyint1 + 1; i < keyint2; i++ {
			k9, k8, k7, k6 := GetStrKeys(i)
			idmap.AddKey(k9, k8, k7, k6, block)
		}
	}
}

func GetKeys(id int) (int, int, int, int) {
	key9 := math.Floor(float64(id) / Power9)
	currentid := id - int(key9*Power9)
	key8 := math.Floor(float64(currentid) / Power8)
	currentid = currentid - int(key8*Power8)
	key7 := math.Floor(float64(currentid) / Power7)
	currentid = currentid - int(key7*Power7)
	key6 := math.Floor(float64(currentid) / Power6)
	currentid = currentid - int(key6*Power6)
	return int(key9), int(key8), int(key7), int(key6)
}

func (idmap *IdMap) GetBlock(id int) int {
	k9, k8, k7, k6 := GetKeys(id)

	for _, i := range idmap.IdMap[k9][k8][k7][k6] {
		if i.IdRange[0] <= id && i.IdRange[1] >= id {
			return i.Position
		}
	}
	return 0
}

func NewIdMap() *IdMap {
	return &IdMap{IdMap: map[int]map[int]map[int]map[int][]*LazyPrimitiveBlock{}}
}
