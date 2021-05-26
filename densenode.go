package osm

import (
	m "github.com/flywave/go-mapbox/tileid"

	"github.com/murphy214/pbf"
)

type Config struct {
	Granularity int
	LatOffset   int
	LonOffset   int
}

type Node struct {
	Tags  map[string]string
	Point []float64
}

type DenseNode struct {
	NodeMap     map[int]*Node
	DenseInfo   int
	KeyValue    int
	BoundingBox m.Extrema
	Tags        []uint32
	Buf         *pbf.PBF
}

func NewConfig() Config {
	return Config{Granularity: 100, LatOffset: 0, LonOffset: 0}
}

type tagUnpacker struct {
	stringTable []string
	keysVals    []int32
	index       int
}

func (tu *tagUnpacker) next() map[string]string {
	tags := make(map[string]string)
	var key, val string
	for tu.index < len(tu.keysVals) {
		keyID := int(tu.keysVals[tu.index])
		tu.index++
		if keyID == 0 {
			break
		}

		valID := int(tu.keysVals[tu.index])
		tu.index++

		if len(tu.stringTable) > keyID {

			key = tu.stringTable[keyID]
		}
		if len(tu.stringTable) > valID {
			val = tu.stringTable[valID]
		}
		if (len(tu.stringTable) > keyID) && (len(tu.stringTable) > valID) {
			tags[key] = val
		}

	}
	return tags
}

func (prim *PrimitiveBlock) NewDenseNode() *DenseNode {
	var tu *tagUnpacker

	densenode := &DenseNode{NodeMap: map[int]*Node{}}
	var idpbf, latpbf, longpbf *pbf.PBF
	key, val := prim.Buf.ReadKey()
	if key == 1 && val == 2 {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		idpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 5 && val == 2 {
		size := prim.Buf.ReadVarint()
		densenode.DenseInfo = prim.Buf.Pos
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 8 && val == 2 {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		latpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 9 && val == 2 {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		longpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 10 && val == 2 {
		densenode.KeyValue = prim.Buf.Pos
		tags := prim.Buf.ReadPackedInt32()
		tu = &tagUnpacker{prim.StringTable, tags, 0}
		key, val = prim.Buf.ReadKey()
	}

	var id, lat, long int
	west, south, east, north := 180.0, 90.0, -180.0, -90.0
	var pt []float64

	for i := 0; i < 8000 || prim.Buf.Pos < prim.Buf.Length; i++ {
		tags := tu.next()
		id = id + int(idpbf.ReadSVarint())
		lat = lat + int(latpbf.ReadSVarint())
		long = long + int(longpbf.ReadSVarint())
		pt = []float64{
			(float64(prim.Config.LonOffset+(long*prim.Config.Granularity)) * 1e-9),
			(float64(prim.Config.LatOffset+(lat*prim.Config.Granularity)) * 1e-9),
		}

		densenode.NodeMap[id] = &Node{Point: pt, Tags: tags}

		x, y := pt[0], pt[1]

		if x < west {
			west = x
		} else if x > east {
			east = x
		}

		if y < south {
			south = y
		} else if y > north {
			north = y
		}
	}

	bds := m.Extrema{N: north, S: south, E: east, W: west}
	densenode.BoundingBox = bds
	densenode.Buf = prim.Buf

	return densenode
}

func (d *Decoder) NewDenseNodeMap(lazy *LazyPrimitiveBlock) map[int][]float64 {
	prim := NewPrimitiveBlockLazy(pbf.NewPBF(d.ReadDataPos(lazy.FilePos)))
	prim.Buf.Pos = prim.GroupIndex[0]

	var idpbf, latpbf, longpbf *pbf.PBF
	key, val := prim.Buf.ReadKey()
	if key == 1 && val == 2 {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		idpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 5 && val == 2 {
		size := prim.Buf.ReadVarint()
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 8 && val == 2 {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		latpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 9 && val == 2 {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		longpbf = pbf.NewPBF(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadKey()
	}
	if key == 10 && val == 2 {
		prim.Buf.ReadPackedInt32()
		key, val = prim.Buf.ReadKey()
	}

	var id, lat, long int
	var pt []float64
	nodemap := map[int][]float64{}
	for i := 0; i < 8000 && idpbf.Pos < idpbf.Length; i++ {
		id = id + int(idpbf.ReadSVarint())
		lat = lat + int(latpbf.ReadSVarint())
		long = long + int(longpbf.ReadSVarint())
		pt = []float64{
			(float64(prim.Config.LonOffset+(long*prim.Config.Granularity)) * 1e-9),
			(float64(prim.Config.LatOffset+(lat*prim.Config.Granularity)) * 1e-9),
		}

		nodemap[id] = pt
	}
	return nodemap
}

func LazyDenseNode(pbfval *pbf.PBF) (int, int, bool) {
	var idpbf *pbf.PBF
	key, val := pbfval.ReadKey()
	var startid, endid int

	if key == 1 && val == 2 {
		size := pbfval.ReadVarint()
		endpos := pbfval.Pos + size
		idpbf = pbf.NewPBF(pbfval.Pbf[pbfval.Pos:endpos])
		id := 0
		for i := 0; i < 8000 && idpbf.Pos < idpbf.Length; i++ {
			id = id + int(idpbf.ReadSVarint())
			if i == 0 {
				startid = id
			}
		}
		endid = id
		pbfval.Pos = endpos
		key, val = pbfval.ReadKey()
	}
	if key == 5 && val == 2 {
		size := pbfval.ReadVarint()
		pbfval.Pos += size
		key, val = pbfval.ReadKey()
	}
	if key == 8 && val == 2 {
		size := pbfval.ReadVarint()
		pbfval.Pos += size
		key, val = pbfval.ReadKey()
	}
	if key == 9 && val == 2 {
		size := pbfval.ReadVarint()
		pbfval.Pos += size
		key, val = pbfval.ReadKey()
	}
	if key == 10 && val == 2 {
		startpos := pbfval.Pos
		pbfval.ReadPackedInt32()
		sizevals := pbfval.Pos - startpos
		return startid, endid, 8002 != sizevals
	}

	return 0, 0, false
}
