package osm

type Diff struct {
	Create bool
	Modify bool
	Delete bool
	Node   *Node
	Way    *Way
	Rel    *Relation
}
