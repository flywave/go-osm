package mapping

import "strconv"

type Key string
type Value string

var defaultRanks map[string]int

func init() {
	defaultRanks = map[string]int{
		"minor":          3,
		"road":           3,
		"unclassified":   3,
		"residential":    3,
		"tertiary_link":  3,
		"tertiary":       4,
		"secondary_link": 3,
		"secondary":      5,
		"primary_link":   3,
		"primary":        6,
		"trunk_link":     3,
		"trunk":          8,
		"motorway_link":  3,
		"motorway":       9,
	}
}

func DefaultWayZOrder(val string, elem *osm.Element, geom *geom.Geometry, match Match) interface{} {
	var z int
	layer, _ := strconv.ParseInt(elem.Tags["layer"], 10, 64)
	z += int(layer) * 10

	rank := defaultRanks[match.Value]

	if rank == 0 {
		if _, ok := elem.Tags["railway"]; ok {
			rank = 7
		}
	}
	z += rank

	tunnel := elem.Tags["tunnel"]
	if tunnel == "true" || tunnel == "yes" || tunnel == "1" {
		z -= 10
	}
	bridge := elem.Tags["bridge"]
	if bridge == "true" || bridge == "yes" || bridge == "1" {
		z += 10
	}

	return z
}
