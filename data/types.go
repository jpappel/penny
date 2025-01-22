package data

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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

func (c Comment) Hash() string {
	str := fmt.Sprint(c.Id, c.Content, c.Hidden, c.Deleted, c.Posted.UTC().Unix(), c.NumChildren)
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}

type PennyDB struct {
	Db *sql.DB // public for testing purposes
}
