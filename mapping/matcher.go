package mapping

func (m *Mapping) pointMatcher() (NodeMatcher, error) {
	mappings := make(TagTableMapping)
	m.mappings(PointTable, mappings)
	filters := make(tableElementFilters)
	m.addFilters(filters)
	m.addTypedFilters(PointTable, filters)
	tables, err := m.tables(PointTable)
	return &tagMatcher{
		mappings:   mappings,
		filters:    filters,
		tables:     tables,
		matchAreas: false,
	}, err
}

func (m *Mapping) lineStringMatcher() (WayMatcher, error) {
	mappings := make(TagTableMapping)
	m.mappings(LineStringTable, mappings)
	filters := make(tableElementFilters)
	m.addFilters(filters)
	m.addTypedFilters(LineStringTable, filters)
	tables, err := m.tables(LineStringTable)
	return &tagMatcher{
		mappings:   mappings,
		filters:    filters,
		tables:     tables,
		matchAreas: false,
	}, err
}

func (m *Mapping) polygonMatcher() (RelWayMatcher, error) {
	mappings := make(TagTableMapping)
	m.mappings(PolygonTable, mappings)
	filters := make(tableElementFilters)
	m.addFilters(filters)
	m.addTypedFilters(PolygonTable, filters)
	relFilters := make(tableElementFilters)
	m.addRelationFilters(PolygonTable, relFilters)
	tables, err := m.tables(PolygonTable)
	return &tagMatcher{
		mappings:   mappings,
		filters:    filters,
		tables:     tables,
		relFilters: relFilters,
		matchAreas: true,
	}, err
}

func (m *Mapping) relationMatcher() (RelationMatcher, error) {
	mappings := make(TagTableMapping)
	m.mappings(RelationTable, mappings)
	filters := make(tableElementFilters)
	m.addFilters(filters)
	m.addTypedFilters(PolygonTable, filters)
	m.addTypedFilters(RelationTable, filters)
	relFilters := make(tableElementFilters)
	m.addRelationFilters(RelationTable, relFilters)
	tables, err := m.tables(RelationTable)
	return &tagMatcher{
		mappings:   mappings,
		filters:    filters,
		tables:     tables,
		relFilters: relFilters,
		matchAreas: true,
	}, err
}

func (m *Mapping) relationMemberMatcher() (RelationMatcher, error) {
	mappings := make(TagTableMapping)
	m.mappings(RelationMemberTable, mappings)
	filters := make(tableElementFilters)
	m.addFilters(filters)
	m.addTypedFilters(RelationMemberTable, filters)
	relFilters := make(tableElementFilters)
	m.addRelationFilters(RelationMemberTable, relFilters)
	tables, err := m.tables(RelationMemberTable)
	return &tagMatcher{
		mappings:   mappings,
		filters:    filters,
		tables:     tables,
		relFilters: relFilters,
		matchAreas: true,
	}, err
}

type NodeMatcher interface {
	MatchNode(node *osm.Node) []Match
}

type WayMatcher interface {
	MatchWay(way *osm.Way) []Match
}

type RelationMatcher interface {
	MatchRelation(rel *osm.Relation) []Match
}

type RelWayMatcher interface {
	WayMatcher
	RelationMatcher
}

type Match struct {
	Key   string
	Value string
	Table DestTable
}

func (m *Match) Row(elem *osm.Element, geom *geom.Geometry) []interface{} {
	return m.builder.MakeRow(elem, geom, *m)
}

func (m *Match) MemberRow(rel *osm.Relation, member *osm.Member, geom *geom.Geometry) []interface{} {
	return m.builder.MakeMemberRow(rel, member, geom, *m)
}

type tagMatcher struct {
	mappings   TagTableMapping
	tables     map[string]*rowBuilder
	filters    tableElementFilters
	relFilters tableElementFilters
	matchAreas bool
}

func (tm *tagMatcher) MatchNode(node *osm.Node) []Match {
	return tm.match(node.Tags, false, false)
}

func (tm *tagMatcher) MatchWay(way *osm.Way) []Match {
	if tm.matchAreas { // match way as polygon
		if way.IsClosed() {
			if way.Tags["area"] == "no" {
				return nil
			}
			return tm.match(way.Tags, true, false)
		}
	} else { // match way as linestring
		if way.IsClosed() {
			if way.Tags["area"] == "yes" {
				return nil
			}
			return tm.match(way.Tags, true, false)
		}
		return tm.match(way.Tags, false, false)
	}
	return nil
}

func (tm *tagMatcher) MatchRelation(rel *osm.Relation) []Match {
	return tm.match(rel.Tags, true, true)
}

type orderedMatch struct {
	Match
	order int
}

func (tm *tagMatcher) match(tags osm.Tags, closed bool, relation bool) []Match {
	tables := make(map[DestTable]orderedMatch)

	addTables := func(k, v string, tbls []orderedDestTable) {
		for _, t := range tbls {
			this := orderedMatch{
				Match: Match{
					Key:     k,
					Value:   v,
					Table:   t.DestTable,
					builder: tm.tables[t.Name],
				},
				order: t.order,
			}
			if other, ok := tables[t.DestTable]; ok {
				if other.order < this.order {
					this = other
				}
			}
			tables[t.DestTable] = this
		}
	}

	if values, ok := tm.mappings[Key("__any__")]; ok {
		addTables("__any__", "__any__", values["__any__"])
	}

	for k, v := range tags {
		values, ok := tm.mappings[Key(k)]
		if ok {
			if tbls, ok := values["__any__"]; ok {
				addTables(k, v, tbls)
			}
			if tbls, ok := values[Value(v)]; ok {
				addTables(k, v, tbls)
			}
		}
	}
	var matches []Match
	for t, match := range tables {
		filters, ok := tm.filters[t.Name]
		filteredOut := false
		if ok {
			for _, filter := range filters {
				if !filter(tags, Key(match.Key), closed) {
					filteredOut = true
					break
				}
			}
		}
		if relation && !filteredOut {
			filters, ok := tm.relFilters[t.Name]
			if ok {
				for _, filter := range filters {
					if !filter(tags, Key(match.Key), closed) {
						filteredOut = true
						break
					}
				}
			}
		}

		if !filteredOut {
			matches = append(matches, match.Match)
		}
	}
	return matches
}
