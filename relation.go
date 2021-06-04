package osm

import "github.com/flywave/go-pbf"

func (d *Decoder) ReadRelationsLazy(lazy *LazyPrimitiveBlock) map[int]int {
	prim := d.CreatePrimitiveBlock(lazy)
	prim.Buf.Pos = prim.GroupIndex[0]
	mymap := map[int]int{}

	for prim.Buf.Pos < prim.GroupIndex[1] {
		prim.Buf.ReadTag()
		endpos2 := prim.Buf.Pos + prim.Buf.ReadVarint()

		key, val := prim.Buf.ReadTag()

		if key == RELATION_ID && val == pbf.Varint {
			prim.Buf.ReadUInt64()
			key, val = prim.Buf.ReadTag()
		}
		if key == RELATION_KEYS {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			key, _ = prim.Buf.ReadTag()
		}
		if key == RELATION_VALS {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			key, _ = prim.Buf.ReadTag()
		}
		if key == RELATION_INFO {
			size := prim.Buf.ReadVarint()
			prim.Buf.Pos += size
			key, _ = prim.Buf.ReadTag()
		}
		if key == RELATION_ROLES_SID {
			size := prim.Buf.ReadVarint()
			endpos := prim.Buf.Pos + size
			prim.Buf.Pos = endpos
			key, _ = prim.Buf.ReadTag()
		}
		if key == RELATION_MEMIDS {
			size := prim.Buf.ReadVarint()
			endpos := prim.Buf.Pos + size
			var x int
			for prim.Buf.Pos < endpos {
				x += int(prim.Buf.ReadSVarint())
				mymap[d.WayIdMap.GetBlock(x)] = 0
			}
		}
		prim.Buf.Pos = endpos2
	}
	return mymap
}
