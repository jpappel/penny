package data

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"iter"
	"time"
)

type PennyDB struct {
	Db *sql.DB // public for testing purposes
}

type Comment struct {
	Id          int
	Content     string
	Hidden      bool
	Deleted     bool
	Posted      time.Time
	NumChildren int
	Children    []Comment
}

type CommentForest []Comment

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

// breadth first iterator over forest
func (cf CommentForest) BFS() iter.Seq[*Comment] {
	return func(yield func(*Comment) bool) {
		queue := make([]*Comment, 0, 5)

		for i := range cf {
			queue = append(queue, &cf[i])
		}

		var c *Comment
		for len(queue) != 0 {
			c = queue[0]
			if !yield(c) {
				return
			}

			queue = queue[1:]

			for i := range c.Children {
				queue = append(queue, &c.Children[i])
			}

		}
	}
}

// depth first iterator over forest
func (cf CommentForest) DFS() iter.Seq[*Comment] {
	return func(yield func(*Comment) bool) {
		stack := make([]*Comment, 0, 5)

		for i := range cf {
			stack = append(stack, &cf[i])
		}

		var c *Comment
		var last int
		for len(stack) != 0 {
			last = len(stack) - 1
			c = stack[last]
			if !yield(c) {
				return
			}

			stack = stack[:last]

			for i := range c.Children {
				stack = append(stack, &c.Children[i])
			}
		}
	}
}
