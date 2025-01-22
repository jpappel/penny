package data

import (
	"context"
	"database/sql"
	"time"
)

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

func (p PennyDB) GetPageCommentsById(ctx context.Context, pageId int) ([]Comment, error) {
	now := ctx.Value("now").(int64)
	result, err := p.Db.QueryContext(ctx, `SELECT id, hiddenTime, deletedTime, postedTime, content, children
    FROM Comments
    JOIN Descendants ON Comments.id = Descendants.commentId
    WHERE pageId = ?`, pageId)
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

func (p PennyDB) GetPageRootComments(ctx context.Context, pageUrl string) ([]Comment, error) {
	now := ctx.Value("now").(int64)
	// PERF: this query is very ugly and should be written with less joins :)
	result, err := p.Db.QueryContext(ctx, `SELECT Comments.id, hiddenTime, deletedTime, postedTime, content, children
    FROM Comments
    JOIN Pages on Comments.pageId = Pages.id
    JOIN Descendants ON Comments.id = Descendants.commentId
    JOIN Relations ON Comments.id = Relations.childId
    WHERE url = ? AND parentId IS NULL`, pageUrl)
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
func (p PennyDB) GetCommentChildren(ctx context.Context, comment *Comment) error {
	if comment.NumChildren == 0 {
		comment.Children = make([]Comment, 0, 5)
		return nil
	}

	now := ctx.Value("now").(int64)
	if comment.Children == nil {
		comment.Children = make([]Comment, 0, comment.NumChildren)
	}

	result, err := p.Db.QueryContext(ctx, `SELECT id, hiddenTime, deletedTime, postedTime, content, children
    FROM Comments
    JOIN Descendants ON Comments.id = Descendants.commentId
    JOIN Relations ON Comments.id = Relations.childId
    WHERE parentId = ?`, comment.Id)
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
