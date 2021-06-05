package osm

import (
	"math"
	"testing"

	"github.com/flywave/go-geom"
)

type TestWriter struct {
	FeatureWriter
	cache []geom.Feature
}

func (w *TestWriter) WriteFeature(feature *geom.Feature) error {
	if w.cache == nil {
		w.cache = make([]geom.Feature, 0, 1024)
	}
	w.cache = append(w.cache, *feature)
	return nil
}

func TestDecoder(t *testing.T) {
	w := &TestWriter{}
	MakeOutputWriter("./testdata/sample.pbf", w, int(math.MaxInt64))

	if len(w.cache) == 0 {
		t.FailNow()
	}
}
