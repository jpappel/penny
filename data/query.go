package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type SortPaginate struct {
	Order      string
	Descending bool
	Limit      int
	Offset     int
}

// Parse a comment from a sql row
func parseComment(result *sql.Rows, unixTime int64) (*Comment, error) {
	comment := new(Comment)
	var hiddenTime sql.NullInt64
	var deletedTime sql.NullInt64
	var parentId sql.NullInt64
	var postedTime int64
	if err := result.Scan(&comment.Id, &parentId, &hiddenTime, &deletedTime, &postedTime, &comment.Content, &comment.NumChildren); err != nil {
		return nil, err
	}

	if parentId.Valid {
		comment.ParentId = int(parentId.Int64)
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

func (p PennyDB) GetPageCommentsById(ctx context.Context, pageId int, sp SortPaginate) (*Page, error) {
	now := ctx.Value("now").(int64)

	var pageUrl string
	err := p.Db.QueryRowContext(ctx, "SELECT url FROM Pages WHERE id = ?", pageId).Scan(&pageUrl)
	if err != nil {
		return nil, err
	}

	query := sp.update(`
    SELECT id, parentId, hiddenTime, deletedTime, postedTime, content, children
    FROM Comments
    JOIN Descendants ON Comments.id = Descendants.commentId
    JOIN Relations ON Comments.id = Relations.childId
    WHERE pageId = ?`)

	result, err := p.Db.QueryContext(ctx, query, pageId)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	page := NewPage(nil)
	page.Url = pageUrl
	page.UpdateTime = now

	for result.Next() {
		comment, err := parseComment(result, now)
		if err != nil {
			return nil, err
		}

		page.Comments.Store(comment.Id, *comment)
		page.Relations.Append(comment.ParentId, comment.Id)
	}

	return page, nil
}

func (p PennyDB) GetPageComments(ctx context.Context, pageUrl string) (*Page, error) {
	now, ok := ctx.Value("now").(int64)
	if !ok {
		return nil, errors.New("Missing `now` in context")
	}

	result, err := p.Db.QueryContext(ctx, `
    SELECT Comments.id, parentId, hiddenTime, deletedTime, postedTime, content, children
    FROM Comments
    JOIN Pages ON Comments.pageId = Pages.id
    JOIN Relations ON Comments.id = Relations.childId
    JOIN Descendants ON Comments.id = Descendants.commentId
    WHERE url = ?`, pageUrl)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	page := NewPage(nil)
	page.Url = pageUrl
	page.UpdateTime = now

    // HACK: workaround for weired behavior in libsql driver
    //       driver from 20241221181756 fails return correct error on empty resultset
	hasRows := false
	for result.Next() {
		hasRows = true
		comment, err := parseComment(result, now)
		if err != nil {
			return nil, err
		}

		page.Comments.Store(comment.Id, *comment)
		page.Relations.Append(comment.ParentId, comment.Id)
	}

	if !hasRows {
		return nil, sql.ErrNoRows
	}

	return page, nil
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
