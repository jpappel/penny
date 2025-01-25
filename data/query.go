package data

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

type SortPaginate struct {
	Order      string
	Descending bool
	Limit      int
	Offset     int
}

// Parse a comment from a sql row
func parseComment(result *sql.Rows, unixTime int64) (Comment, error) {
	comment := Comment{}
	var hiddenTime sql.NullInt64
	var deletedTime sql.NullInt64
	var postedTime int64
	if err := result.Scan(&comment.Id, &hiddenTime, &deletedTime, &postedTime, &comment.Content, &comment.NumChildren); err != nil {
		return Comment{}, err
	}

	if hiddenTime.Valid {
		comment.Hidden = hiddenTime.Int64 < unixTime
	}

	if deletedTime.Valid {
		comment.Deleted = deletedTime.Int64 < unixTime
	}
	comment.Posted = time.Unix(postedTime, 0)

	return comment, nil
}

// Creates a sorted and or paginated version of a query
func (sp SortPaginate) update(query string) string {
	newQuery := query

	if sp.Order == "id" || sp.Order == "postedTime" {
		orderStr := "ASC"
		if sp.Descending {
			orderStr = "DESC"
		}
		newQuery = fmt.Sprintf("%s\nORDER BY %s %s", newQuery, sp.Order, orderStr)
	}

	if sp.Limit > 0 && sp.Offset >= 0 {
		newQuery = fmt.Sprintf("%s\nLIMIT %d\nOFFSET %d", newQuery, sp.Limit, sp.Offset)
	}

	return newQuery
}

func (p PennyDB) GetPageCommentsById(ctx context.Context, pageId int, sp SortPaginate) ([]Comment, error) {
	now := ctx.Value("now").(int64)

	query := sp.update(`
    SELECT id, hiddenTime, deletedTime, postedTime, content, children
    FROM Comments
    JOIN Descendants ON Comments.id = Descendants.commentId
    WHERE pageId = ?`)

	result, err := p.Db.QueryContext(ctx, query, pageId)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	comments := make([]Comment, 0)
	for result.Next() {
		comment, err := parseComment(result, now)
		if err != nil {
			return comments, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (p PennyDB) GetPageComments(ctx context.Context, pageUrl string) ([]Comment, error) {
	now := ctx.Value("now").(int64)
	result, err := p.Db.QueryContext(ctx, `SELECT Comments.id, hiddenTime, deletedTime, postedTime, content, children
    FROM Comments
    JOIN Pages ON Comments.pageId = Pages.id
    JOIN Descendants ON Comments.id = Descendants.commentId
    WHERE url = ?`, pageUrl)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	comments := make([]Comment, 0)
	for result.Next() {
		comment, err := parseComment(result, now)
		if err != nil {
			return comments, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (p PennyDB) GetPageRootComments(ctx context.Context, pageUrl string, sp SortPaginate) ([]Comment, error) {
	now := ctx.Value("now").(int64)
	// PERF: this query is very ugly and should be written with less joins :)
	query := sp.update(`
    SELECT Comments.id, hiddenTime, deletedTime, postedTime, content, children
    FROM Comments
    JOIN Pages on Comments.pageId = Pages.id
    JOIN Descendants ON Comments.id = Descendants.commentId
    JOIN Relations ON Comments.id = Relations.childId
    WHERE url = ? AND parentId IS NULL`)

	result, err := p.Db.QueryContext(ctx, query, pageUrl)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	comments := make([]Comment, 0)
	for result.Next() {
		comment, err := parseComment(result, now)
		if err != nil {
			return comments, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (p PennyDB) GetCommentById(ctx context.Context, commentId int) (Comment, error) {
	now := ctx.Value("now").(int64)
	row := p.Db.QueryRowContext(ctx, `SELECT id, hiddenTime, deletedTime, postedTime, content, children
    FROM Comments JOIN Descendants ON Comments.id = Descendants.commentId WHERE id = ?`, commentId)

	comment := Comment{}
	var hiddenTime sql.NullInt64
	var deletedTime sql.NullInt64
	var postedTime int64
	if err := row.Scan(&comment.Id, &hiddenTime, &deletedTime, &postedTime, &comment.Content, &comment.NumChildren); err != nil {
		return Comment{}, err
	}

	if hiddenTime.Valid {
		comment.Hidden = hiddenTime.Int64 < now
	}

	if deletedTime.Valid {
		comment.Deleted = deletedTime.Int64 < now
	}
	comment.Posted = time.Unix(postedTime, 0)

	return comment, nil
}

// Adds a comments children to it
func (p PennyDB) GetCommentChildren(ctx context.Context, comment *Comment, sp SortPaginate) error {
	if comment.NumChildren == 0 {
		comment.Children = make([]Comment, 0, 5)
		return nil
	}

	now := ctx.Value("now").(int64)
	if comment.Children == nil {
		comment.Children = make([]Comment, 0, comment.NumChildren)
	}

	query := sp.update(`
    SELECT id, hiddenTime, deletedTime, postedTime, content, children
    FROM Comments
    JOIN Descendants ON Comments.id = Descendants.commentId
    JOIN Relations ON Comments.id = Relations.childId
    WHERE parentId = ?`)

	result, err := p.Db.QueryContext(ctx, query, comment.Id)
	if err != nil {
		return err
	}
	defer result.Close()

	for result.Next() {
		childComment, err := parseComment(result, now)
		if err != nil {
			return err
		}
		comment.Children = append(comment.Children, childComment)
	}

	return nil
}

func getChildren(ctx context.Context, queue chan *Comment, stmt *sql.Stmt) error {
	now := ctx.Value("now").(int64)
	comment := <-queue
	rows, err := stmt.QueryContext(ctx, comment.Id)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		childComment, err := parseComment(rows, now)
		if err != nil {
			return err
		}
		comment.Children = append(comment.Children, childComment)
		queue <- &comment.Children[len(comment.Children)-1]
	}

	return nil
}

// Add children to a comment recursively in a breadth first order up to a relative depth (exclusive)
func (p PennyDB) BFSGetCommentChildren(ctx context.Context, rootComment *Comment, depth int, sp SortPaginate) error {
	tx, err := p.Db.BeginTx(ctx, new(sql.TxOptions))
	if err != nil {
		return err
	}

	var rootDepth int
	row := tx.QueryRowContext(ctx, "SELECT depth FROM Relations WHERE childId = ?", rootComment.Id)
	if err := row.Scan(&rootDepth); err != nil {
		tx.Rollback()
		return err
	}

	var numDescendants int64
	row = tx.QueryRowContext(ctx, `
    WITH RECURSIVE
    ddsc(childId, depth) AS (
        VALUES(?, ?)
        UNION ALL
        SELECT Relations.childId, Relations.depth
        FROM Relations
        JOIN ddsc ON Relations.parentId = ddsc.childId
    )
    SELECT COUNT(*) - 1
    FROM ddsc
    WHERE depth < ?`, rootComment.Id, rootDepth, rootDepth+depth)
	if err := row.Scan(&numDescendants); err != nil {
		tx.Rollback()
		return err
	}

	// a number of descendants is returned even for invalid comment Id's
	// added to throw the expected error
	if numDescendants < 0 {
		tx.Rollback()
		return sql.ErrNoRows
	} else if numDescendants == 0 {
		tx.Commit()
		return nil
	}

	var wg sync.WaitGroup
	// PERF: should test different channel sizes
	queue := make(chan *Comment, 256)
	queue <- rootComment

	query := sp.update(`
    SELECT id, hiddenTime, deletedTime, postedTime, content, children
    FROM Comments
    JOIN Descendants ON Comments.id = Descendants.commentId
    JOIN Relations ON Comments.id = Relations.childId
    WHERE parentId = ?`)

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	// add root Comment back
	for range numDescendants + 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := getChildren(ctx, queue, stmt); err != nil {
				// TODO: handle panic
				panic(err)
			}
		}()
	}

	wg.Wait()
	tx.Commit()

	return nil
}
