package osm

import (
	"fmt"
	"time"
)

type Tags map[string]string

func (t *Tags) String() string {
	return fmt.Sprintf("%v", (map[string]string)(*t))
}

type Element struct {
	ID       int64
	Tags     Tags
	Metadata *Metadata
}

type Metadata struct {
	UserID    int32
	UserName  string
	Version   int32
	Timestamp time.Time
	Changeset int64
}

type Node struct {
	Element
	Lat  float64
	Long float64
}

type Way struct {
	Element
	Refs  []int64
	Nodes []Node
}

func (w *Way) IsClosed() bool {
	return len(w.Refs) >= 4 && w.Refs[0] == w.Refs[len(w.Refs)-1]
}

type MemberType int

const (
	NodeMember     MemberType = 0
	WayMember                 = 1
	RelationMember            = 2
)

type Relation struct {
	Element
	Members []Member
}

type Member struct {
	ID      int64
	Type    MemberType
	Role    string
	Way     *Way
	Node    *Node
	Element *Element
}
