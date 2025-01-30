package data_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jpappel/penny/data"
)

const MaxInt64 int64 = 1<<63 - 1

type CommentsTestCase struct {
	name        string
	expected    *data.Page
	expectedErr error
	setup       func(string) data.PennyDB
	runner      func(data.PennyDB) (*data.Page, error)
}

func (tc CommentsTestCase) Test(t *testing.T) {
	dir := t.TempDir()
	pdb := tc.setup(fmt.Sprintf("file:%s/%s.db", dir, tc.name))
	p, err := tc.runner(pdb)

	if err != tc.expectedErr {
		t.Fatalf("Unexpected error: wanted `%v` got `%v`\n", tc.expectedErr, err)
	}

	if p == nil && tc.expected == nil {
		return
	} else if p == nil || tc.expected == nil {
		t.Fatalf("Recieved nil pages: expected %p, result %p\n", tc.expected, p)
	}

	pLen, eLen := p.Len(), tc.expected.Len()
	minLen := min(pLen, eLen)
	if pLen != eLen {
		t.Errorf("Received different number of comments: wanted %d, got %d\n", eLen, pLen)
		t.Logf("Comparing up to %d comments\n", minLen)
	}

	for i := range minLen {
		c, e := p.Comments[i], tc.expected.Comments[i]
		if cHash, eHash := c.Hash(), e.Hash(); cHash != eHash {
			t.Errorf("Recieved a different comment hash than expected: wanted %s, got %s\n", eHash, cHash)

			if c.Id != e.Id {
				t.Logf("Different Id's: wanted %d, got %d\n", e.Id, c.Id)
			}
			if c.Deleted != e.Deleted {
				t.Logf("Different Deletion status: wanted %t, got %t\n", e.Deleted, c.Deleted)
			}
			if c.Hidden != e.Hidden {
				t.Logf("Different Hidden status: wanted %t, got %t\n", e.Hidden, c.Hidden)
			}
			if !c.Posted.Equal(e.Posted) {
				t.Logf("Different Posted Time: wanted %v, got %v\n", e.Posted, c.Posted)
			}

			cReplies, eReplies := len(c.Replies), len(e.Replies)
			minReplies := min(cReplies, eReplies)
			if cReplies != eReplies {
				t.Logf("Different Number of replies: wanted %d, got %d\n", eReplies, cReplies)
				t.Logf("Comparing up to %d replies\n", minReplies)
			}
			for j := range minReplies {
				if c.Replies[j] != e.Replies[j] {
					t.Logf("Different reply than expected: wanted %d, got %d\n", e.Replies[j], c.Replies[j])
				}
			}

			if c.Content != e.Content {
				t.Logf("Different Content's:\nwanted:\n%s\n<END>\ngot:\n%s\n<END>\n", e.Content, c.Content)
			}
		}
	}
}

// single comment with a single user
func singleComment(connStr string) data.PennyDB {
	pdb := data.New(connStr)

	tx, err := pdb.Db.Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec("INSERT INTO Users(email, provider, name) VALUES (?,?,?)", "a@z.com", "github", "A Z")
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec("INSERT INTO Pages(url) VALUES (?)", "apples")
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec("INSERT INTO Comments(userId, pageId, postedTime, content) VALUES (1,1,0,?)", "pie")
	if err != nil {
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	return pdb
}

var singleCommentPage *data.Page

// singleComment but hidden
func hiddenComment(connStr string) data.PennyDB {
	pdb := singleComment(connStr)

	tx, err := pdb.Db.Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec("UPDATE Comments SET hiddenTime = 1")
	if err != nil {
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	return pdb
}

func deletedComment(connStr string) data.PennyDB {
	pdb := singleComment(connStr)
	tx, err := pdb.Db.Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec("UPDATE Comments SET deletedTime = 2, content = ? WHERE id = 1", "")
	if err != nil {
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	return pdb
}

// 3 comment chain with 2 authors
// root -> commenter -> root commenter response
func nestedCommentChain(connStr string) data.PennyDB {
	pdb := data.New(connStr)

	tx, err := pdb.Db.Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec(`
    INSERT INTO Users(email, provider, name)
    VALUES (?,?,?), (?,?,?)
    `, "a@z.com", "github", "A Z", "b@y.org", "google", "B Y")
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec("INSERT INTO Pages(url) VALUES (?)", "peaches")
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec(`
    INSERT INTO Comments(userId, pageId, postedTime, content)
    VALUES (?,?,?,?), (?,?,?,?), (?,?,?,?)`,
		1, 1, 0, "cobbler",
		2, 1, 1, "with",
		1, 1, 2, "icecream")
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec(`
    INSERT INTO Replies(parentId, childId)
    VALUES (?,?),(?,?)`,
		1, 2,
		2, 3)
	if err != nil {
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	return pdb
}

var nestedCommentChainPage *data.Page

// multiple root comments from multiple authors
/*
page (the)
|-- 1 "first"  (A Z #1) t0
|----- 4 "letter" (B Y #2) t3
|-------- 8 "of the english alphabet descends from proto-sinatic script" (A Z #1) t5
|-------- 9 "is an inverted bull" (C X #3) t5
|----- 5 "animal" (C X #3) t3
|-- 2 "second" (B Y #2) t1
|----- 6 "ammendment" (A Z #1) t4
|-------- 7 "of the US constitution is the right to bear arms" (C X #3) t5
|-- 3 "last"   (C X #3) t2
|----- 10 "christmas" (B Y #2) t7
|----- 11 "I gave you my heart" (A Z #1) t8
|----- 12 "but then the very next day" (C X #3) t9
|----- 13 "you gave it away" (A Z #1) t10
*/
func commentForest(connStr string) data.PennyDB {
	pdb := data.New(connStr)

	tx, err := pdb.Db.Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec(`
    INSERT INTO USERS(email, provider, name)
    VALUES (?,?,?), (?,?,?), (?,?,?)`,
		"a@z.com", "github", "A Z",
		"b@y.org", "google", "B Y",
		"c@x.net", "twitter", "C X")
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec("INSERT INTO Pages(url) VALUES (?)", "the")
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare("INSERT INTO Comments(userId, pageId, postedTime, content) VALUES (?,1,?,?)")
	if err != nil {
		panic(err)
	}

	comments := []struct {
		userId     int64
		postedTime int64
		content    string
	}{
		{1, 0, "first"},
		{2, 1, "second"},
		{3, 2, "last"},
		{2, 3, "letter"},
		{3, 3, "animal"},
		{1, 4, "ammendment"},
		{3, 5, "of the US constitution is the right to bear arms"},
		{1, 5, "of the english alphabet descends from proto-sinatic script"},
		{3, 5, "is an inverted bull"},
		{2, 7, "christmas"},
		{1, 8, "I gave you my heart"},
		{3, 9, "but then the very next day"},
		{1, 10, "you gave it away"},
	}
	for _, c := range comments {
		_, err = stmt.Exec(c.userId, c.postedTime, c.content)
		if err != nil {
			panic(err)
		}
	}
	stmt.Close()

	stmt, err = tx.Prepare("INSERT INTO Replies(parentId, childId) VALUES (?,?)")
	if err != nil {
		panic(err)
	}

	relations := []struct {
		parent int
		self   int
	}{
		{1, 4}, {1, 5}, {4, 8}, {4, 9},
		{2, 6}, {6, 7},
		{3, 10}, {3, 11}, {3, 12}, {3, 13},
	}
	for _, r := range relations {
		_, err = stmt.Exec(r.parent, r.self)
		if err != nil {
			panic(err)
		}
	}
	stmt.Close()

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	return pdb
}

var commentForestPage *data.Page

func init() {
	singleCommentPage = &data.Page{
		Url:        "apples",
		UpdateTime: time.Unix(MaxInt64, 0),
		Comments:   []data.Comment{{1, "pie", false, false, time.Unix(0, 0), nil}},
	}

	nestedCommentChainPage = &data.Page{
		Url:        "peaches",
		UpdateTime: time.Unix(MaxInt64, 0),
		Comments: []data.Comment{
			{1, "cobbler", false, false, time.Unix(0, 0), []int{2}},
			{2, "with", false, false, time.Unix(1, 0), []int{3}},
			{3, "icecream", false, false, time.Unix(2, 0), nil},
		}}

	commentForestPage = &data.Page{
		Url:        "the",
		UpdateTime: time.Unix(MaxInt64, 0),
		Comments: []data.Comment{
			{Id: 1, Content: "first", Posted: time.Unix(0, 0), Replies: []int{4, 5}},
			{Id: 2, Content: "second", Posted: time.Unix(1, 0), Replies: []int{6}},
			{Id: 3, Content: "last", Posted: time.Unix(2, 0), Replies: []int{10, 11, 12, 13}},
			{Id: 4, Content: "letter", Posted: time.Unix(3, 0), Replies: []int{8, 9}},
			{Id: 5, Content: "animal", Posted: time.Unix(3, 0), Replies: nil},
			{Id: 6, Content: "ammendment", Posted: time.Unix(4, 0), Replies: []int{7}},
			{Id: 7, Content: "of the US constitution is the right to bear arms", Posted: time.Unix(5, 0), Replies: nil},
			{Id: 8, Content: "of the english alphabet descends from proto-sinatic script", Posted: time.Unix(5, 0), Replies: nil},
			{Id: 9, Content: "is an inverted bull", Posted: time.Unix(5, 0), Replies: nil},
			{Id: 10, Content: "christmas", Posted: time.Unix(7, 0), Replies: nil},
			{Id: 11, Content: "I gave you my heart", Posted: time.Unix(8, 0), Replies: nil},
			{Id: 12, Content: "but then the very next day", Posted: time.Unix(9, 0), Replies: nil},
			{Id: 13, Content: "you gave it away", Posted: time.Unix(10, 0), Replies: nil},
		}}
}
