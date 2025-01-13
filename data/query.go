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
	if err := result.Scan(&comment.Id, &hiddenTime, &deletedTime, &postedTime, &(comment.Content)); err != nil {
		return Comment{}, nil
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
	now := time.Now().UTC().Unix()
	result, err := p.db.QueryContext(ctx, `SELECT id, hiddenTime, deletedTime, postedTime, content
    FROM Comments
    WHERE pageId = ? AND (deletedTime > ? OR deletedTime IS NULL)`, pageId, now)
	if err != nil {
		panic(err)
	}
	defer result.Close()

	comments := make([]Comment, 0)
	for result.Next() {
		comment, err := parseComment(result, now)
		if err != nil {
			panic(err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (p PennyDB) GetPageComments(ctx context.Context, pageUrl string) ([]Comment, error) {
	now := time.Now().UTC().Unix()
	result, err := p.db.QueryContext(ctx, `SELECT Comments.id, hiddenTime, deletedTime, postedTime, content
    FROM Comments
    JOIN Pages ON Comments.pageId = Pages.id
    WHERE url = ? AND (deletedTime > ? OR deletedTime IS NULL)`, pageUrl, now)
	if err != nil {
		panic(err)
	}
	defer result.Close()

	comments := make([]Comment, 0)
	for result.Next() {
		comment, err := parseComment(result, now)
		if err != nil {
			panic(err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}
