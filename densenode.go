package osm

import (
	m "github.com/flywave/go-mapbox/tileid"

	"github.com/flywave/go-pbf"
)

type Config struct {
	Granularity int
	LatOffset   int
	LonOffset   int
}

type DenseNode struct {
	NodeMap     map[int]*Node
	DenseInfo   int
	KeyValue    int
	BoundingBox m.Extrema
	Tags        []uint32
	Buf         *pbf.Reader
}

func NewConfig() Config {
	return Config{Granularity: 100, LatOffset: 0, LonOffset: 0}
}

type tagUnpacker struct {
	stringTable [][]byte
	keysVals    []int32
	index       int
}

func (tu *tagUnpacker) next() Tags {
	tags := make(Tags)
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

			key = string(tu.stringTable[keyID])
		}
		if len(tu.stringTable) > valID {
			val = string(tu.stringTable[valID])
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
	var idpbf, latpbf, longpbf *pbf.Reader
	key, val := prim.Buf.ReadTag()
	if key == DENSE_NODE_ID && val == pbf.Bytes {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		idpbf = pbf.NewReader(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadTag()
	}
	if key == DENSE_NODE_INFO && val == pbf.Bytes {
		size := prim.Buf.ReadVarint()
		densenode.DenseInfo = prim.Buf.Pos
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadTag()
	}
	if key == DENSE_NODE_LAT && val == pbf.Bytes {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		latpbf = pbf.NewReader(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadTag()
	}
	if key == DENSE_NODE_LON && val == pbf.Bytes {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		longpbf = pbf.NewReader(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadTag()
	}
	if key == DENSE_NODE_KEYS_VALS && val == pbf.Bytes {
		densenode.KeyValue = prim.Buf.Pos
		tags := prim.Buf.ReadPackedInt32()
		tu = &tagUnpacker{prim.StringTable, tags, 0}
		key, val = prim.Buf.ReadTag()
	}

	var id, lat, long int
	west, south, east, north := 180.0, 90.0, -180.0, -90.0
	var pt []float64

	for i := 0; i < 8000 || prim.Buf.Pos < prim.Buf.Length; i++ {
		tags := tu.next()
		id = id + int(idpbf.ReadSVarint())
		lat = lat + int(latpbf.ReadSVarint())
		long = long + int(longpbf.ReadSVarint())

		flong := (float64(prim.Config.LonOffset+(long*prim.Config.Granularity)) * 1e-9)
		flat := (float64(prim.Config.LatOffset+(lat*prim.Config.Granularity)) * 1e-9)

		densenode.NodeMap[id] = &Node{Lat: flat, Long: flong, Element: Element{ID: int64(id), Tags: tags}}

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
	prim := NewPrimitiveBlockLazy(pbf.NewReader(d.ReadDataPos(lazy.FilePos)))
	prim.Buf.Pos = prim.GroupIndex[0]

	var idpbf, latpbf, longpbf *pbf.Reader
	key, val := prim.Buf.ReadTag()
	if key == DENSE_NODE_ID && val == pbf.Bytes {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		idpbf = pbf.NewReader(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadTag()
	}
	if key == DENSE_NODE_INFO && val == pbf.Bytes {
		size := prim.Buf.ReadVarint()
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadTag()
	}
	if key == DENSE_NODE_LAT && val == pbf.Bytes {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		latpbf = pbf.NewReader(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadTag()
	}
	if key == DENSE_NODE_LON && val == pbf.Bytes {
		size := prim.Buf.ReadVarint()
		endpos := prim.Buf.Pos + size
		longpbf = pbf.NewReader(prim.Buf.Pbf[prim.Buf.Pos:endpos])
		prim.Buf.Pos += size
		key, val = prim.Buf.ReadTag()
	}
	if key == DENSE_NODE_KEYS_VALS && val == pbf.Bytes {
		prim.Buf.ReadPackedInt32()
		key, val = prim.Buf.ReadTag()
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

func LazyDenseNode(pbfval *pbf.Reader) (int, int, bool) {
	var idpbf *pbf.Reader
	key, val := pbfval.ReadTag()
	var startid, endid int

	if key == DENSE_NODE_ID && val == pbf.Bytes {
		size := pbfval.ReadVarint()
		endpos := pbfval.Pos + size
		idpbf = pbf.NewReader(pbfval.Pbf[pbfval.Pos:endpos])
		id := 0
		for i := 0; i < 8000 && idpbf.Pos < idpbf.Length; i++ {
			id = id + int(idpbf.ReadSVarint())
			if i == 0 {
				startid = id
			}
		}
		endid = id
		pbfval.Pos = endpos
		key, val = pbfval.ReadTag()
	}
	if key == DENSE_NODE_INFO && val == pbf.Bytes {
		size := pbfval.ReadVarint()
		pbfval.Pos += size
		key, val = pbfval.ReadTag()
	}
	if key == DENSE_NODE_LAT && val == pbf.Bytes {
		size := pbfval.ReadVarint()
		pbfval.Pos += size
		key, val = pbfval.ReadTag()
	}
	if key == DENSE_NODE_LON && val == pbf.Bytes {
		size := pbfval.ReadVarint()
		pbfval.Pos += size
		key, val = pbfval.ReadTag()
	}
	if key == DENSE_NODE_KEYS_VALS && val == pbf.Bytes {
		startpos := pbfval.Pos
		pbfval.ReadPackedInt32()
		sizevals := pbfval.Pos - startpos
		return startid, endid, 8002 != sizevals
	}

	return 0, 0, false
}
