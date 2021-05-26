package osm

import (
	"github.com/murphy214/pbf"
)

type PrimitiveBlock struct {
	StringTable []string
	GroupIndex  [2]int
	GroupType   int
	Buf         *pbf.PBF
	Config      Config
}

func NewPrimitiveBlock(pbfval *pbf.PBF) *PrimitiveBlock {
	primblock := &PrimitiveBlock{}
	key, val := pbfval.ReadKey()
	if key == 1 && val == 2 {
		size := pbfval.ReadVarint()
		endpos := pbfval.Pos + size
		for pbfval.Pos < endpos {
			_, _ = pbfval.ReadKey()
			primblock.StringTable = append(primblock.StringTable, pbfval.ReadString())
		}

		pbfval.Pos = endpos
		key, val = pbfval.ReadKey()
	}
	if key == 2 && val == 2 {
		endpos := pbfval.Pos + pbfval.ReadVarint()
		grouptype, _ := pbfval.ReadKey()
		if grouptype == 2 {
			pbfval.ReadVarint()
		} else if grouptype == 3 {
			pbfval.Pos -= 1
		}

		primblock.GroupIndex = [2]int{pbfval.Pos, endpos}

		primblock.GroupType = int(grouptype)
		pbfval.Pos = endpos
		key, val = pbfval.ReadKey()
	}
	if key == 100 {
		primblock.Config = NewConfig()
	}

	primblock.Buf = pbfval

	return primblock
}

type LazyPrimitiveBlock struct {
	Type     string
	IdRange  [2]int
	FilePos  [2]int
	BufPos   [2]int
	Position int
	TagsBool bool
}

func ReadLazyPrimitiveBlock(pbfval *pbf.PBF) LazyPrimitiveBlock {
	var lazyblock LazyPrimitiveBlock
	key, val := pbfval.ReadKey()
	if key == 1 && val == 2 {
		size := pbfval.ReadVarint()
		endpos := pbfval.Pos + size
		pbfval.Pos = endpos
		key, val = pbfval.ReadKey()
	}
	if key == 2 && val == 2 {
		endpos := pbfval.Pos + pbfval.ReadVarint()
		grouptype, _ := pbfval.ReadKey()
		if grouptype == 1 {
			lazyblock.Type = "Nodes"
			pbfval.Pos -= 1
		} else if grouptype == 2 {
			pbfval.ReadVarint()
			lazyblock.Type = "DenseNodes"
		} else if grouptype == 3 {
			lazyblock.Type = "Ways"
			pbfval.Pos -= 1
		} else if grouptype == 4 {
			lazyblock.Type = "Relations"
			pbfval.Pos -= 1
		} else if grouptype == 5 {
			lazyblock.Type = "Changesets"
			pbfval.Pos -= 1
		}
		lazyblock.BufPos = [2]int{pbfval.Pos, endpos}
		if lazyblock.Type == "DenseNodes" {
			start, end, boolval := LazyDenseNode(pbfval)
			lazyblock.IdRange = [2]int{start, end}
			lazyblock.TagsBool = boolval
		} else if lazyblock.Type == "Ways" {
			start, end := LazyWayRange(pbfval)
			lazyblock.IdRange = [2]int{start, end}
		}

	}

	return lazyblock
}

func NewPrimitiveBlockLazy(pbfval *pbf.PBF) *PrimitiveBlock {
	primblock := &PrimitiveBlock{}

	key, val := pbfval.ReadKey()
	if key == 1 && val == 2 {

		size := pbfval.ReadVarint()
		endpos := pbfval.Pos + size
		pbfval.Pos = endpos
		key, val = pbfval.ReadKey()
	}
	if key == 2 && val == 2 {
		endpos := pbfval.Pos + pbfval.ReadVarint()
		grouptype, _ := pbfval.ReadKey()
		if grouptype == 2 {
			pbfval.ReadVarint()
		} else if grouptype == 3 {
			pbfval.Pos -= 1
		}

		primblock.GroupIndex = [2]int{pbfval.Pos, endpos}

		primblock.GroupType = int(grouptype)
		pbfval.Pos = endpos
		key, val = pbfval.ReadKey()
	}
	if key == 100 {
		primblock.Config = NewConfig()
	}

	primblock.Buf = pbfval

	return primblock
}
