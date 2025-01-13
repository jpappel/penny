package data

import (
	"database/sql"
	"fmt"
	"time"
)

type Comment struct {
	Id          int
	Content     string
	Hidden      bool
	Deleted     bool
	Posted      time.Time
	NumChildren int
	Children    []Comment
}

func (c Comment) String() string {
	formatStr := "Comment %d: hidden[%t] deleted[%t]\nPosted (UTC) %s\n%d Children\n---\n%s"
	return fmt.Sprintf(formatStr,
		c.Id, c.Hidden, c.Deleted, c.Posted.String(), c.NumChildren, c.Content)
}

type PennyDB struct {
	db *sql.DB
}
