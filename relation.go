package osm

func (d *Decoder) ReadRelationsLazy(lazy *LazyPrimitiveBlock) map[int]int {
	prim := d.CreatePrimitiveBlock(lazy)
	prim.Buf.Pos = prim.GroupIndex[0]
	mymap := map[int]int{}

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
			prim.Buf.Pos = endpos
			key, _ = prim.Buf.ReadKey()
		}

		if key == 9 {
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
