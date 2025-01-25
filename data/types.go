package data

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"iter"
	"slices"
	"strings"
	"time"
)

type PennyDB struct {
	Db *sql.DB // public for testing purposes
}


// TODO: consider using an adjacency list with sync.Map isntead of slice
//
//	something like `Page := sync.Map[int]struct{comment Comment, children []int}
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
type CommentSorter func(c1, c2 Comment) int

func ByNone(c1, c2 Comment) int {
	return 0
}

func ById(c1, c2 Comment) int {
	if c1.Id < c2.Id {
		return -1
	} else if c1.Id > c2.Id {
		return 1
	} else {
		return 0
	}
}

func ByPosted(c1, c2 Comment) int {
	return c1.Posted.Compare(c2.Posted)
}

// Sorts a comments children
func (c Comment) Sort(eq CommentSorter) {
	slices.SortFunc(c.Children, eq)
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

func (cf CommentForest) Len() int {
	size := 0
	for range cf.BFS(ByNone) {
		size += 1
	}

	return size
}

// breadth first iterator over forest
// May mutate order of children if not already sorted
func (cf CommentForest) BFS(By CommentSorter) iter.Seq[*Comment] {
	return func(yield func(*Comment) bool) {
		queue := make([]*Comment, 0, 5)

		slices.SortFunc(cf, By)

		for i := range cf {
			queue = append(queue, &cf[i])
		}

		var c *Comment
		for len(queue) != 0 {
			c = queue[0]
			if !yield(c) {
				return
			}

			// PERF: consider using two pointers instead of reslicing
			queue = queue[1:]

			slices.SortFunc(c.Children, By)

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

func (cf CommentForest) String() string {
	var b strings.Builder
	for c := range cf.BFS(ByNone) {
		fmt.Fprintf(&b, "%d %p --> [", c.Id, c)
		for _, child := range c.Children {
			fmt.Fprintf(&b, "%d,", child.Id)
		}
		fmt.Fprint(&b, "]\n")
	}
	return b.String()
}
