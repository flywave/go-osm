package osm

import "github.com/flywave/go-geom"

type FeatureWriter interface {
	WriteFeature(feature *geom.Feature) error
}
