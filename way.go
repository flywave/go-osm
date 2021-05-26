package osm

import (
	"github.com/murphy214/pbf"
)

func (d *Decoder) CreatePrimitiveBlock(lazy *LazyPrimitiveBlock) *PrimitiveBlock {
	return &PrimitiveBlock{Buf: pbf.NewPBF(d.ReadDataPos(lazy.FilePos)), GroupIndex: lazy.BufPos, GroupType: 3}
}

func (d *Decoder) ReadWaysLazy(lazy *LazyPrimitiveBlock, idmap *IdMap) map[int]string {
	prim := d.CreatePrimitiveBlock(lazy)
	prim.Buf.Pos = prim.GroupIndex[0]
	mymap := map[int]string{}

	for prim.Buf.Pos < prim.GroupIndex[1] {
		prim.Buf.ReadKey()
		endpos2 := prim.Buf.Pos + prim.Buf.ReadVarint()

		key, val := prim.Buf.ReadKey()
		if key == 1 && val == 0 {
			prim.Buf.ReadUInt64()
			key, val = prim.Buf.ReadKey()
		}
		if key == 2 {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			key, _ = prim.Buf.ReadKey()
		}
		if key == 3 {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			key, _ = prim.Buf.ReadKey()
		}

		if key == 4 {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			key, _ = prim.Buf.ReadKey()
		}

		if key == 8 {
			size := prim.Buf.ReadVarint()
			endpos := prim.Buf.Pos + size
			var x int
			for prim.Buf.Pos < endpos {
				x += int(prim.Buf.ReadSVarint())
				position := idmap.GetBlock(x)
				mymap[position] = ""
			}

			prim.Buf.Pos += size + 1
		}
		prim.Buf.Pos = endpos2
	}
	return mymap
}

func (d *Decoder) ReadWaysLazyList(lazy *LazyPrimitiveBlock, ids []int) map[int][]int {
	idmap := map[int]string{}
	for _, i := range ids {
		idmap[i] = ""
	}

	prim := d.CreatePrimitiveBlock(lazy)
	prim.Buf.Pos = prim.GroupIndex[0]
	mymap := map[int][]int{}
	var boolval bool
	var id int
	for prim.Buf.Pos < prim.GroupIndex[1] {
		prim.Buf.ReadKey()
		endpos2 := prim.Buf.Pos + prim.Buf.ReadVarint()

		key, val := prim.Buf.ReadKey()
		if key == 1 && val == 0 {
			id = int(prim.Buf.ReadUInt64())
			_, boolval = idmap[id]
			key, val = prim.Buf.ReadKey()
		}
		if key == 2 {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			key, _ = prim.Buf.ReadKey()
		}
		if key == 3 {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			key, _ = prim.Buf.ReadKey()
		}

		if key == 4 {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			key, _ = prim.Buf.ReadKey()
		}

		if key == 8 {
			size := prim.Buf.ReadVarint()
			endpos := prim.Buf.Pos + size
			if boolval {
				var x int
				var xlist []int
				for prim.Buf.Pos < endpos {
					x += int(prim.Buf.ReadSVarint())
					xlist = append(xlist, x)
				}
				prim.Buf.Pos += size + 1
				mymap[id] = xlist
			} else {
				prim.Buf.Pos = endpos
			}

		}
		prim.Buf.Pos = endpos2
	}
	return mymap
}

func LazyWayRange(pbfval *pbf.PBF) (int, int) {
	var start, pos, id int
	for pbfval.Pos < pbfval.Length {
		pbfval.ReadKey()
		endpos2 := pbfval.Pos + pbfval.ReadVarint()

		key, val := pbfval.ReadKey()

		if key == 1 && val == 0 {
			id = int(pbfval.ReadUInt64())
			if pos == 0 {
				start = id
			}
			key, val = pbfval.ReadKey()
		}
		if key == 2 {
			size := pbfval.ReadVarint()
			pbfval.Pos += size
			key, _ = pbfval.ReadKey()
		}
		if key == 3 {
			size := pbfval.ReadVarint()
			pbfval.Pos += size
			key, _ = pbfval.ReadKey()
		}

		if key == 4 {
			size := pbfval.ReadVarint()
			pbfval.Pos += size
			key, _ = pbfval.ReadKey()
		}

		if key == 8 {
			size := pbfval.ReadVarint()
			endpos := pbfval.Pos + size
			pbfval.Pos = endpos
		}
		pbfval.Pos = endpos2
		pos++
	}

	return start, id
}

func (d *Decoder) SyncWaysNodeMap(lazy *LazyPrimitiveBlock, idmap *IdMap) {
	keymap := d.ReadWaysLazy(lazy, idmap)
	keylist := make([]int, len(keymap))
	i := 0
	for k := range keymap {
		keylist[i] = k
		i++
	}
	d.AddUpdates(keylist)
}

func (d *Decoder) SyncWaysNodeMapMultiple(lazys []*LazyPrimitiveBlock, idmap *IdMap) {
	keymap := map[int]string{}
	c := make(chan map[int]string)
	current := 0
	for pos, lazy := range lazys {
		go func(lazy *LazyPrimitiveBlock) {
			c <- d.ReadWaysLazy(lazy, idmap)
		}(lazy)
		current++

		if pos%10 == 1 || len(lazys)-1 == pos {
			for i := 0; i < current; i++ {
				tempmap := <-c
				for k, v := range tempmap {
					keymap[k] = v
				}
			}
			current = 0
		}
	}

	keylist := make([]int, len(keymap))
	i := 0
	for k := range keymap {
		keylist[i] = k
		i++
	}
	d.AddUpdates(keylist)
}
