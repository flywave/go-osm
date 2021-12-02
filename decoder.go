package osm

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	osmpbf "github.com/flywave/go-osm/osmpbf"

	m "github.com/flywave/go-mapbox/tileid"

	"github.com/gogo/protobuf/proto"

	"github.com/flywave/go-pbf"
)

const (
	MaxBlobHeaderSize = 64 * 1024
	MaxBlobSize       = 32 * 1024 * 1024
)

var (
	parseCapabilities = map[string]bool{
		"OsmSchema-V0.6":        true,
		"DenseNodes":            true,
		"HistoricalInformation": true,
	}
)

const (
	osmHeaderType = "OSMHeader"
	osmDataType   = "OSMData"
)

type Header struct {
	Bounds               *m.Extrema
	RequiredFeatures     []string
	OptionalFeatures     []string
	WritingProgram       string
	Source               string
	ReplicationTimestamp time.Time
	ReplicationSeqNum    uint64
	ReplicationBaseURL   string
}

type iPair struct {
	Offset int64
	Blob   *osmpbf.Blob
	Err    error
}

type Decoder struct {
	Header      *Header
	r           io.Reader
	bytesRead   int64
	Count       int
	DenseNodes  map[int]*LazyPrimitiveBlock
	Ways        map[int]*LazyPrimitiveBlock
	Relations   map[int]*LazyPrimitiveBlock
	Nodes       map[int]*LazyPrimitiveBlock
	IdMap       *IdMap
	WayIdMap    *IdMap
	NodeMap     *NodeMap
	RelationMap map[int]string
	Limit       int
	Writer      FeatureWriter
	WriteBool   bool
	TotalMemory int
	cancel      func()
	wg          sync.WaitGroup

	inputs []chan<- iPair

	m sync.Mutex

	cOffset int64
	cIndex  int
	f       *os.File
}

func NewDecoder(f *os.File, limit int, writer FeatureWriter) *Decoder {
	return &Decoder{
		r:           f,
		f:           f,
		DenseNodes:  map[int]*LazyPrimitiveBlock{},
		Ways:        map[int]*LazyPrimitiveBlock{},
		Relations:   map[int]*LazyPrimitiveBlock{},
		Nodes:       map[int]*LazyPrimitiveBlock{},
		NodeMap:     NewNodeMap(limit),
		IdMap:       NewIdMap(),
		WayIdMap:    NewIdMap(),
		RelationMap: map[int]string{},
		Writer:      writer,
		Limit:       limit,
		WriteBool:   true,
	}
}

func (dec *Decoder) Close() error {
	if dec.cancel != nil {
		dec.cancel()
		dec.wg.Wait()
	}
	return nil
}

func (dec *Decoder) ReadDataPos(pos [2]int) []byte {
	buf := make([]byte, int(pos[1]-pos[0]))
	dec.f.ReadAt(buf, int64(pos[0]))

	blob := &osmpbf.Blob{}
	if err := proto.Unmarshal(buf, blob); err != nil {
		fmt.Println(err)
	}

	data, err := GetData(blob)
	if err != nil {
		fmt.Println(err)
	}
	dec.TotalMemory += len(data)
	return data
}

func (dec *Decoder) ReadBlock(lazyprim LazyPrimitiveBlock) *osmpbf.PrimitiveBlock {
	primblock := &osmpbf.PrimitiveBlock{}
	err := proto.Unmarshal(dec.ReadDataPos(lazyprim.FilePos), primblock)
	if err != nil {
		fmt.Println(err)
	}
	return primblock
}

func ReadDecoder(f *os.File, limit int, writer FeatureWriter) *Decoder {
	d := NewDecoder(f, limit, writer)
	sizeBuf := make([]byte, 4)
	headerBuf := make([]byte, MaxBlobHeaderSize)
	blobBuf := make([]byte, MaxBlobSize)

	_, blob, _, index := d.ReadFileBlock(sizeBuf, headerBuf, blobBuf)
	header, err := DecodeOSMHeader(blob)
	if err != nil {
		fmt.Println(err)
	}
	d.Header = header
	fi, _ := f.Stat()
	filesize := int(fi.Size()) / 1000000

	boolval := true
	var oldsize int64
	c := make(chan *LazyPrimitiveBlock)
	increment := 0
	for boolval {
		d.Count++
		_, blob, _, index = d.ReadFileBlock(sizeBuf, headerBuf, blobBuf)
		count := d.Count
		go func(blob *osmpbf.Blob, index [2]int, count int, c chan *LazyPrimitiveBlock) {
			if blob != nil {
				bytevals, err := GetData(blob)
				if err != nil {
					fmt.Println(err)
				}
				primblock := ReadLazyPrimitiveBlock(pbf.NewReader(bytevals))
				primblock.Position = count
				primblock.FilePos = index
				c <- &primblock

			} else {
				c <- &LazyPrimitiveBlock{}
			}

		}(blob, index, count, c)

		increment++

		if increment == 1000 || d.bytesRead == oldsize {
			for myc := 0; myc < increment; myc++ {
				primblock := <-c
				switch primblock.Type {
				case "DenseNodes":
					d.DenseNodes[primblock.Position] = primblock
					d.IdMap.AddBlock(primblock)
				case "Ways":
					d.Ways[primblock.Position] = primblock
					d.WayIdMap.AddBlock(primblock)
				case "Relations":
					d.Relations[primblock.Position] = primblock
				case "Nodes":
					d.Nodes[primblock.Position] = primblock
				}
			}
			increment = 0
		}
		if d.bytesRead == oldsize {
			boolval = false
		}
		oldsize = d.bytesRead
		fmt.Printf("\r[%dmb/%dmb] concurrent preliminary read with %d fileblocks total", d.bytesRead/1000000, filesize, d.Count)
	}

	return d
}

func MakeOutputWriter(infilename string, writer FeatureWriter, limit int) {
	f, _ := os.Open(infilename)
	d := ReadDecoder(f, limit, writer)
	d.ProcessFile()
}

func (dec *Decoder) ReadFileBlock(sizeBuf, headerBuf, blobBuf []byte) (*osmpbf.BlobHeader, *osmpbf.Blob, error, [2]int) {
	blobHeaderSize, err := dec.ReadBlobHeaderSize(sizeBuf)
	if err != nil {
		return nil, nil, err, [2]int{0, 0}
	}
	headerBuf = headerBuf[:blobHeaderSize]
	blobHeader, err := dec.ReadBlobHeader(headerBuf)
	if err != nil {
		return nil, nil, err, [2]int{0, 0}
	}

	blobBuf = blobBuf[:blobHeader.GetDatasize()]
	blob, err := dec.ReadBlob(blobHeader, blobBuf)
	if err != nil {
		return nil, nil, err, [2]int{0, 0}
	}

	dec.bytesRead += 4 + int64(blobHeaderSize)
	index := [2]int{int(dec.bytesRead), int(dec.bytesRead) + int(blobHeader.GetDatasize())}

	dec.bytesRead += int64(blobHeader.GetDatasize())

	return blobHeader, blob, nil, index
}

func (dec *Decoder) ReadBlobHeaderSize(buf []byte) (uint32, error) {
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return 0, err
	}

	size := binary.BigEndian.Uint32(buf)
	if size >= MaxBlobHeaderSize {
		return 0, errors.New("BlobHeader size >= 64Kb")
	}
	return size, nil
}

func (dec *Decoder) ReadBlobHeader(buf []byte) (*osmpbf.BlobHeader, error) {
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return nil, err
	}

	blobHeader := &osmpbf.BlobHeader{}
	if err := proto.Unmarshal(buf, blobHeader); err != nil {
		return nil, err
	}

	if blobHeader.GetDatasize() >= MaxBlobSize {
		return nil, errors.New("Blob size >= 32Mb")
	}
	return blobHeader, nil
}

func (dec *Decoder) ReadBlob(blobHeader *osmpbf.BlobHeader, buf []byte) (*osmpbf.Blob, error) {
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return nil, err
	}

	blob := &osmpbf.Blob{}
	if err := proto.Unmarshal(buf, blob); err != nil {
		return nil, err
	}
	return blob, nil
}

func GetData(blob *osmpbf.Blob) ([]byte, error) {
	switch {
	case blob.Raw != nil:
		return blob.GetRaw(), nil

	case blob.ZlibData != nil:
		r, err := zlib.NewReader(bytes.NewReader(blob.GetZlibData()))
		if err != nil {
			return nil, err
		}
		buf := make([]byte, blob.GetRawSize())
		_, err = io.ReadFull(r, buf)

		if len(buf) != int(blob.GetRawSize()) {
			return nil, fmt.Errorf("raw blob data size %d but expected %d", len(buf), blob.GetRawSize())
		}

		return buf, nil
	default:
		return nil, errors.New("unknown blob data")
	}
}

func DecodeOSMHeader(blob *osmpbf.Blob) (*Header, error) {
	data, err := GetData(blob)
	if err != nil {
		return nil, err
	}

	headerBlock := &osmpbf.HeaderBlock{}
	if err := proto.Unmarshal(data, headerBlock); err != nil {
		return nil, err
	}

	requiredFeatures := headerBlock.GetRequiredFeatures()
	for _, feature := range requiredFeatures {
		if !parseCapabilities[feature] {
			return nil, fmt.Errorf("parser does not have %s capability", feature)
		}
	}

	header := &Header{
		RequiredFeatures:   headerBlock.GetRequiredFeatures(),
		OptionalFeatures:   headerBlock.GetOptionalFeatures(),
		WritingProgram:     headerBlock.GetWritingprogram(),
		Source:             headerBlock.GetSource(),
		ReplicationBaseURL: headerBlock.GetOsmosisReplicationBaseUrl(),
		ReplicationSeqNum:  uint64(headerBlock.GetOsmosisReplicationSequenceNumber()),
	}

	if headerBlock.OsmosisReplicationTimestamp != 0 {
		header.ReplicationTimestamp = time.Unix(headerBlock.OsmosisReplicationTimestamp, 0).UTC()
	}
	if headerBlock.Bbox != nil {
		header.Bounds = &m.Extrema{
			W: 1e-9 * float64(headerBlock.Bbox.Left),
			E: 1e-9 * float64(headerBlock.Bbox.Right),
			S: 1e-9 * float64(headerBlock.Bbox.Bottom),
			N: 1e-9 * float64(headerBlock.Bbox.Top),
		}
	}

	return header, nil
}
