package data_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jpappel/penny/data"
)

const MaxInt64 int64 = 1<<63 - 1

type CommentsTestCase struct {
	name        string
	expected    []data.Comment
	expectedErr error
	setup       func(string) data.PennyDB
	runner      func(data.PennyDB) ([]data.Comment, error)
}

func (c CommentsTestCase) Test(t *testing.T) {
	dir := t.TempDir()
	p := c.setup(fmt.Sprintf("file:%s/%s.db", dir, c.name))
	comments, err := c.runner(p)

	if err != c.expectedErr {
		t.Fatalf("Unexpected error in GetPageCommentsById: wanted %v got %v\n", c.expectedErr, err)
	}

	if len(comments) != len(c.expected) {
		t.Errorf("Recieved a different number of comments than expected: wanted %d got %d\n",
			len(c.expected), len(comments))
	}

	expHashes := make(map[string]bool, len(c.expected))
	for _, expComment := range c.expected {
		expHashes[expComment.Hash()] = true
	}

	for _, comment := range comments {
		hash := comment.Hash()
		_, ok := expHashes[hash]

		if !ok {
			t.Error("Recieved an unexpected comment:", comment)
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

	_, err = tx.Exec("INSERT INTO Relations(parentId, childId, depth) VALUES (NULL, 1, 0)")
	if err != nil {
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	return pdb
}

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
    INSERT INTO Relations(parentId, childId, depth)
    VALUES (?,?,?),(?,?,?),(?,?,?)`,
		nil, 1, 0,
		1, 2, 1,
		2, 3, 2)
	if err != nil {
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	return pdb
}

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

	_, err = tx.Exec(`
    INSERT INTO Relations(parentId, childId, depth)
    VALUES (NULL,?,0), (NULL,?,0), (NULL,?,0)`, 1, 2, 3)
	if err != nil {
		panic(err)
	}

	stmt, err = tx.Prepare("INSERT INTO Relations(parentId, childId, depth) VALUES (?,?,?)")
	if err != nil {
		panic(err)
	}

	relations := []struct {
		parent int
		self   int
		depth  int
	}{
		{1, 4, 1}, {1, 5, 1}, {4, 8, 2}, {4, 9, 2},
		{2, 6, 1}, {6, 7, 2},
		{3, 10, 1}, {3, 11, 1}, {3, 12, 1}, {3, 13, 1},
	}
	for _, r := range relations {
		_, err = stmt.Exec(r.parent, r.self, r.depth)
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

func TestGetPageCommentsById(t *testing.T) {
	testCases := []CommentsTestCase{
		{"MissingPage",
			nil,
			nil,
			singleComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageCommentsById(ctx, -1)
			}},
		{"SingleComment",
			[]data.Comment{
				{1, "pie", false, false, time.Unix(0, 0), 0, nil},
			},
			nil,
			singleComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageCommentsById(ctx, 1)
			}},
		{"NestedCommentChain",
			[]data.Comment{
				{1, "cobbler", false, false, time.Unix(0, 0), 1, nil},
				{2, "with", false, false, time.Unix(1, 0), 1, nil},
				{3, "icecream", false, false, time.Unix(2, 0), 0, nil},
			},
			nil,
			nestedCommentChain,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageCommentsById(ctx, 1)
			},
		},
		{"CommentForest",
			[]data.Comment{
				{1, "first", false, false, time.Unix(0, 0), 2, nil},
				{2, "second", false, false, time.Unix(1, 0), 1, nil},
				{3, "last", false, false, time.Unix(2, 0), 4, nil},
				{4, "letter", false, false, time.Unix(3, 0), 2, nil},
				{5, "animal", false, false, time.Unix(3, 0), 0, nil},
				{6, "ammendment", false, false, time.Unix(4, 0), 1, nil},
				{7, "of the US constitution is the right to bear arms", false, false, time.Unix(5, 0), 0, nil},
				{8, "of the english alphabet descends from proto-sinatic script", false, false, time.Unix(5, 0), 0, nil},
				{9, "is an inverted bull", false, false, time.Unix(5, 0), 0, nil},
				{10, "christmas", false, false, time.Unix(7, 0), 0, nil},
				{11, "I gave you my heart", false, false, time.Unix(8, 0), 0, nil},
				{12, "but then the very next day", false, false, time.Unix(9, 0), 0, nil},
				{13, "you gave it away", false, false, time.Unix(10, 0), 0, nil},
			},
			nil,
			commentForest,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageCommentsById(ctx, 1)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, testCase.Test)
	}
}

func TestGetPageComments(t *testing.T) {
	testCases := []CommentsTestCase{
		{"MissingPage",
			nil,
			nil,
			singleComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageComments(ctx, "I do not exist")
			}},
		{"SingleComment",
			[]data.Comment{
				{1, "pie", false, false, time.Unix(0, 0), 0, nil},
			},
			nil,
			singleComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageComments(ctx, "apples")
			}},
		{"NestedCommentChain",
			[]data.Comment{
				{1, "cobbler", false, false, time.Unix(0, 0), 1, nil},
				{2, "with", false, false, time.Unix(1, 0), 1, nil},
				{3, "icecream", false, false, time.Unix(2, 0), 0, nil},
			},
			nil,
			nestedCommentChain,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageComments(ctx, "peaches")
			},
		},
		{"CommentForest",
			[]data.Comment{
				{1, "first", false, false, time.Unix(0, 0), 2, nil},
				{2, "second", false, false, time.Unix(1, 0), 1, nil},
				{3, "last", false, false, time.Unix(2, 0), 4, nil},
				{4, "letter", false, false, time.Unix(3, 0), 2, nil},
				{5, "animal", false, false, time.Unix(3, 0), 0, nil},
				{6, "ammendment", false, false, time.Unix(4, 0), 1, nil},
				{7, "of the US constitution is the right to bear arms", false, false, time.Unix(5, 0), 0, nil},
				{8, "of the english alphabet descends from proto-sinatic script", false, false, time.Unix(5, 0), 0, nil},
				{9, "is an inverted bull", false, false, time.Unix(5, 0), 0, nil},
				{10, "christmas", false, false, time.Unix(7, 0), 0, nil},
				{11, "I gave you my heart", false, false, time.Unix(8, 0), 0, nil},
				{12, "but then the very next day", false, false, time.Unix(9, 0), 0, nil},
				{13, "you gave it away", false, false, time.Unix(10, 0), 0, nil},
			},
			nil,
			commentForest,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageComments(ctx, "the")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, testCase.Test)
	}
}

func TestGetPageRootComments(t *testing.T) {
	testCases := []CommentsTestCase{
		{"MissingPage",
			nil,
			nil,
			singleComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageRootComments(ctx, "I do not exist")
			}},
		{"SingleComment",
			[]data.Comment{
				{1, "pie", false, false, time.Unix(0, 0), 0, nil},
			},
			nil,
			singleComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageRootComments(ctx, "apples")
			}},
		{"NestedCommentChain",
			[]data.Comment{
				{1, "cobbler", false, false, time.Unix(0, 0), 1, nil},
			},
			nil,
			nestedCommentChain,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageRootComments(ctx, "peaches")
			},
		},
		{"CommentForest",
			[]data.Comment{
				{1, "first", false, false, time.Unix(0, 0), 2, nil},
				{2, "second", false, false, time.Unix(1, 0), 1, nil},
				{3, "last", false, false, time.Unix(2, 0), 4, nil},
			},
			nil,
			commentForest,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				return p.GetPageRootComments(ctx, "the")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, testCase.Test)
	}
}

func TestGetCommentsById(t *testing.T) {
	testCases := []CommentsTestCase{
		{"NoValidComment",
			[]data.Comment{{}},
			sql.ErrNoRows,
			singleComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				comment, err := p.GetCommentById(ctx, 100)
				return []data.Comment{comment}, err
			}},
		{"ValidComment",
			[]data.Comment{{1, "pie", false, false, time.Unix(0, 0), 0, nil}},
			nil,
			singleComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				comment, err := p.GetCommentById(ctx, 1)
				return []data.Comment{comment}, err
			}},
		{"HiddenComment",
			[]data.Comment{{1, "pie", true, false, time.Unix(0, 0), 0, nil}},
			nil,
			hiddenComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				comment, err := p.GetCommentById(ctx, 1)
				return []data.Comment{comment}, err
			},
		},
		{"DeletedComment",
			[]data.Comment{{1, "", false, true, time.Unix(0, 0), 0, nil}},
			nil,
			deletedComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				comment, err := p.GetCommentById(ctx, 1)
				return []data.Comment{comment}, err
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, testCase.Test)
	}
}

func TestGetCommentChildren(t *testing.T) {
	testCases := []CommentsTestCase{
		{"NoValidComment",
			[]data.Comment{},
			nil,
			singleComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				comment := data.Comment{}
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				err := p.GetCommentChildren(ctx, &comment)
				return comment.Children, err
			},
		},
		{"NoChildren", // a great tmg song
			[]data.Comment{},
			nil,
			singleComment,
			func(p data.PennyDB) ([]data.Comment, error) {
				comment := data.Comment{1, "pie", false, false, time.Unix(0, 0), 0, nil}
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				err := p.GetCommentChildren(ctx, &comment)
				return comment.Children, err
			},
		},
		{"Children",
			[]data.Comment{
				{10, "christmas", false, false, time.Unix(7, 0), 0, nil},
				{11, "I gave you my heart", false, false, time.Unix(8, 0), 0, nil},
				{12, "but then the very next day", false, false, time.Unix(9, 0), 0, nil},
				{13, "you gave it away", false, false, time.Unix(10, 0), 0, nil},
			},
			nil,
			commentForest,
			func(p data.PennyDB) ([]data.Comment, error) {
				comment := data.Comment{3, "last", false, false, time.Unix(2, 0), 4, nil}
				ctx := context.WithValue(context.Background(), "now", MaxInt64)
				err := p.GetCommentChildren(ctx, &comment)
				return comment.Children, err
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, testCase.Test)
	}
}
