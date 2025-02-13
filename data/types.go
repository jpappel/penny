package data

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type PennyDB struct {
	Db *sql.DB // public for testing purposes
}

type Comment struct {
	Id      int
	Content string
	Hidden  bool
	Deleted bool
	Posted  time.Time
	Replies []int
}

type PageInfo struct {
	Url         string
	UpdateTime  time.Time // TODO: change update time to time of last comment post
	Open        bool
	NumComments int
}

// TODO: put pages into a pool
type Page struct {
	PageInfo
	Comments []Comment
}

func (c Comment) String() string {
	formatStr := "Comment %d: hidden[%t] deleted[%t]\nPosted (UTC) %s\n%d Children\n---\n%s"
	return fmt.Sprintf(formatStr,
		c.Id, c.Hidden, c.Deleted, c.Posted.String(), len(c.Replies), c.Content)
}

func (c Comment) Hash() string {
	str := fmt.Sprint(c.Id, c.Content, c.Hidden, c.Deleted, c.Posted.UTC().Unix(), len(c.Replies), c.Replies)
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}

func (p Page) Len() int {
	return len(p.Comments)
}

func (p Page) String() string {
	var b strings.Builder
	for _, c := range p.Comments {
		fmt.Fprintln(&b, c)
	}
	return b.String()
}

var ErrNoPage error = errors.New("No matching page")
