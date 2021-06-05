package osm

import (
	"time"
)

type Changeset struct {
	ID         int64
	CreatedAt  time.Time
	ClosedAt   time.Time
	Open       bool
	UserID     int32
	UserName   string
	NumChanges int32
	MaxExtent  [4]float64
	Comments   []Comment
	Tags       Tags
}

type Comment struct {
	UserID    int32
	UserName  string
	CreatedAt time.Time
	Text      string
}
