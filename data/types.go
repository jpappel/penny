package data

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"iter"
	"slices"
	"strings"
	"sync"
	"time"
)

type PennyDB struct {
	Db *sql.DB // public for testing purposes
}

type Comment struct {
	Id          int
	ParentId    int
	Content     string
	Hidden      bool
	Deleted     bool
	Posted      time.Time
	NumChildren int
}

// A concurent Map
type SharedAdjList struct {
	Nodes map[int][]int
	sync.RWMutex
}

// TODO: put pages into a pool
type Page struct {
	Url        string
	UpdateTime int64
	Comments   sync.Map
	Relations  SharedAdjList // 0 is used for comments with no root
}

type CommentSorter func(c1, c2 Comment) int

func (c Comment) String() string {
	formatStr := "Comment %d: hidden[%t] deleted[%t]\nPosted (UTC) %s\n%d Children\n---\n%s"
	return fmt.Sprintf(formatStr,
		c.Id, c.Hidden, c.Deleted, c.Posted.String(), c.NumChildren, c.Content)
}

func (c Comment) Hash() string {
	str := fmt.Sprint(c.Id, c.ParentId, c.Content, c.Hidden, c.Deleted, c.Posted.UTC().Unix(), c.NumChildren)
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}

func (p *Page) Len() int {
	return len(p.Relations.Nodes) - 1
}

func (p *Page) BFS() iter.Seq[int] {
	return func(yield func(int) bool) {
		queue := make([]int, 0, p.Len()/2+1)

		p.Relations.RLock()
		defer p.Relations.RUnlock()

		queue = append(queue, p.Relations.Nodes[0]...)
		for len(queue) != 0 {
			id := queue[0]
			if !yield(id) {
				return
			}

			queue = queue[1:]
			queue = append(queue, p.Relations.Nodes[id]...)
		}
	}
}

func (p *Page) DFS() iter.Seq[int] {
	return func(yield func(int) bool) {
		stack := make([]int, 0, p.Len()/2+1)

		p.Relations.RLock()
		defer p.Relations.RUnlock()

		stack = append(stack, p.Relations.Nodes[0]...)
		var top int
		for len(stack) != 0 {
			top = len(stack) - 1
			id := stack[top]
			if !yield(id) {
				return
			}

			stack = stack[:top]
			stack = append(stack, p.Relations.Nodes[id]...)
		}
	}
}

func (p *Page) String() string {
	var b strings.Builder
	for id := range p.BFS() {
		fmt.Fprintln(&b, id, "--->", p.Relations.Nodes[id])
	}
	return b.String()
}

// Append a child to a parents list
func (sal *SharedAdjList) Append(parent int, child int) {
	sal.Lock()
	defer sal.Unlock()

	sal.Nodes[parent] = append(sal.Nodes[parent], child)
	slices.Sort(sal.Nodes[parent])

	if _, ok := sal.Nodes[child]; !ok {
		sal.Nodes[child] = nil
	}
}

// Get the children of a node
func (sal *SharedAdjList) Get(id int) ([]int, bool) {
	sal.RLock()
	defer sal.RUnlock()

	children, ok := sal.Nodes[id]
	return children, ok
}

// create a page from a slice of comments
func NewPage(comments []Comment) *Page {
	page := new(Page)
	page.Relations.Nodes = make(map[int][]int)

	for _, comment := range comments {
		page.Comments.Store(comment.Id, comment)
		page.Relations.Append(comment.ParentId, comment.Id)
	}

	return page
}
