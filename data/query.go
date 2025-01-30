package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SortPaginate struct {
	Order      string
	Descending bool
	Limit      int
	Offset     int
}

// Parse a comment from a sql row
func parseComment(ctx context.Context, row *sql.Rows, stmt *sql.Stmt, unixTime int64) (*Comment, error) {
	comment := new(Comment)
	var hiddenTime sql.NullInt64
	var deletedTime sql.NullInt64
	var postedTime int64
	if err := row.Scan(&comment.Id, &hiddenTime, &deletedTime, &postedTime, &comment.Content); err != nil {
		return nil, err
	}

	if hiddenTime.Valid {
		comment.Hidden = hiddenTime.Int64 < unixTime
	}

	if deletedTime.Valid {
		comment.Deleted = deletedTime.Int64 < unixTime
	}
	comment.Posted = time.Unix(postedTime, 0)

	result, err := stmt.QueryContext(ctx, comment.Id)
	if err == sql.ErrNoRows {
		return comment, nil
	} else if err != nil {
		return comment, err
	}

	var replyId int
	for result.Next() {
		if err := result.Scan(&replyId); err != nil {
			// TODO: handle succesful comment parse but incorrect replies parse
			return comment, err
		}
		comment.Replies = append(comment.Replies, replyId)
	}

	return comment, nil
}

func (p PennyDB) GetPageCommentsById(ctx context.Context, pageId int) (*Page, error) {
	now, ok := ctx.Value("now").(int64)
	if !ok {
		return nil, errors.New("Missing `now` in context")
	}

	var pageUrl string
	err := p.Db.QueryRowContext(ctx, "SELECT url FROM Pages WHERE id = ?", pageId).Scan(&pageUrl)
	if err != nil {
		return nil, err
	}

	query := `
    SELECT id, hiddenTime, deletedTime, postedTime, content
    FROM Comments
    WHERE pageId = ?
    ORDER BY postedTime`

	result, err := p.Db.QueryContext(ctx, query, pageId)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	stmt, err := p.Db.PrepareContext(ctx, `
    SELECT childId
    FROM Replies
    WHERE parentId = ?
    ORDER BY childId`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	page := new(Page)
	page.Url = pageUrl
	page.UpdateTime = time.Unix(now, 0)

	for result.Next() {
		comment, err := parseComment(ctx, result, stmt, now)
		if err != nil {
			return nil, err
		}
		page.Comments = append(page.Comments, *comment)
	}

	return page, nil
}

func (p PennyDB) GetPageComments(ctx context.Context, pageUrl string) (*Page, error) {
	now, ok := ctx.Value("now").(int64)
	if !ok {
		return nil, errors.New("Missing `now` in context")
	}

	result, err := p.Db.QueryContext(ctx, `
    SELECT Comments.id, hiddenTime, deletedTime, postedTime, content
    FROM Comments
    JOIN Pages ON Comments.pageId = Pages.id
    WHERE url = ?
    ORDER BY postedTime`, pageUrl)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	page := new(Page)
	page.Url = pageUrl
	page.UpdateTime = time.Unix(now, 0)

	stmt, err := p.Db.PrepareContext(ctx, `
    SELECT childId
    FROM Replies
    WHERE parentId = ?
    ORDER BY childId`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// HACK: workaround for weired behavior in libsql driver
	//       driver from 20241221181756 fails return correct error on empty resultset
	hasRows := false
	for result.Next() {
		hasRows = true
		comment, err := parseComment(ctx, result, stmt, now)
		if err != nil {
			return nil, err
		}
		page.Comments = append(page.Comments, *comment)
	}

	if !hasRows {
		return nil, ErrNoPage
	}

	return page, nil
}

func (p PennyDB) GetCommentById(ctx context.Context, commentId int) (Comment, error) {
	now, ok := ctx.Value("now").(int64)
	if !ok {
		return Comment{}, errors.New("Missing `now` in context")
	}

	row := p.Db.QueryRowContext(ctx, `
    SELECT id, hiddenTime, deletedTime, postedTime, content
    FROM Comments
    WHERE id = ?`, commentId)

	comment := Comment{}
	var hiddenTime sql.NullInt64
	var deletedTime sql.NullInt64
	var postedTime int64
	if err := row.Scan(&comment.Id, &hiddenTime, &deletedTime, &postedTime, &comment.Content); err != nil {
		return Comment{}, err
	}

	if hiddenTime.Valid {
		comment.Hidden = hiddenTime.Int64 < now
	}

	if deletedTime.Valid {
		comment.Deleted = deletedTime.Int64 < now
	}
	comment.Posted = time.Unix(postedTime, 0)

	result, err := p.Db.QueryContext(ctx, `
    SELECT childId
    FROM Replies
    WHERE parentId = ?
    ORDER BY childId`)
	if err == sql.ErrNoRows {
		return comment, nil
	} else if err != nil {
		return Comment{}, err
	}

	var replyId int
	for result.Next() {
		if err := result.Scan(&replyId); err != nil {
			return Comment{}, err
		}
		comment.Replies = append(comment.Replies, replyId)
	}

	return comment, nil
}
