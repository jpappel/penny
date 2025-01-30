package data_test

import (
	"fmt"
	"iter"
	"slices"
	"testing"

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

	pNext, pStop := iter.Pull(p.BFS())
	eNext, eStop := iter.Pull(tc.expected.BFS())

	pId, pMore := pNext()
	eId, eMore := eNext()
	for pMore && eMore {
		if pId != eId {
			t.Errorf("Recieved a diffent id than expected: wanted %d got %d\n", eId, pId)
		}
		pId, pMore = pNext()
		eId, eMore = eNext()
	}
	pStop()
	eStop()

	if t.Failed() {
		t.Log("Expected:\n\n", tc.expected, "\n")
		t.Log("Recieved:\n\n", p)
	}

	pIds := make([]int, 0, p.Len())
	eIds := make([]int, 0, tc.expected.Len())

	p.Comments.Range(func(k any, _ any) bool {
		id, ok := k.(int)
		if !ok {
			t.Fatal("Recieved unexpected key in comments map", k)
		}
		pIds = append(pIds, id)
		return true
	})
	tc.expected.Comments.Range(func(k any, _ any) bool {
		id, ok := k.(int)
		if !ok {
			t.Fatal("Recieved unexpected key in expected comments map", k)
		}
		eIds = append(eIds, id)
		return true
	})

	slices.Sort(pIds)
	slices.Sort(eIds)
	for i := range pIds {
		if pIds[i] != eIds[i] {
			t.Fatalf("Recieved a different id than expected: wanted %d got %d\n", eIds[i], pIds[i])
		}
	}

	eHashes := make(map[int]string, len(eIds))
	tc.expected.Comments.Range(func(k any, v any) bool {
		id := k.(int)
		comment, ok := v.(data.Comment)
		if !ok {
			t.Fatal("Recieved unexpected value in expected comments map", v)
		}

		eHashes[id] = comment.Hash()
		return true
	})
	p.Comments.Range(func(k any, v any) bool {
		id := k.(int)
		comment, ok := v.(data.Comment)
		if !ok {
			t.Fatal("Recieved unexpected value in comments map", v)
		}

		if cHash, eHash := comment.Hash(), eHashes[id]; cHash != eHash {
			t.Errorf("Recieved a different comment hash than expected: id=%d wanted %s, got %s\n", id, eHash, cHash)
			e, _ := tc.expected.Comments.Load(id)
			eC := e.(data.Comment)
			if comment.ParentId != eC.ParentId {
				t.Logf("\tDiffernt ParentId's: wanted %d, got %d\n", comment.ParentId, eC.ParentId)
			}
			if comment.NumChildren != eC.NumChildren {
				t.Logf("\tDiffernt NumChildren's: wanted %d, got %d\n", comment.NumChildren, eC.NumChildren)
			}
			if comment.Deleted != eC.Deleted {
				t.Logf("\tDiffernt Deletion status: wanted %t, got %t\n", comment.Deleted, eC.Deleted)
			}
			if comment.Hidden != eC.Hidden {
				t.Logf("\tDiffernt Hidden status: wanted %t, got %t\n", comment.Hidden, eC.Hidden)
			}
			if !comment.Posted.Equal(eC.Posted) {
				t.Logf("\tDifferent Posted Time: wanted %v, got %v\n", comment.Posted, eC.Posted)
			}
			if comment.Content != eC.Content {
				t.Logf("\tDifferent Content's:\nwanted:\n%s\n<END>\ngot:\n%s\n<END>", comment.Content, eC.Content)
			}
			// t.Log("Recieved\n", comment)
			// t.Log("\nExpected\n", eComment)
		}

		return true
	})
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
