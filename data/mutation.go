package data

import (
	"context"
	"database/sql"
	"time"
)

func (p PennyDB) PostComment(ctx context.Context, page string, user string, comment Comment, parentId *int64) (int, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	var userId int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM Users WHERE email = ?", user).Scan(&userId)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return -1, err
	} else if err != nil {
		tx.Rollback()
		panic(err)
	}

	var pageId int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM Pages WHERE url = ?", page).Scan(&pageId)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return -1, err
	} else if err != nil {
		tx.Rollback()
		panic(err)
	}

	now := time.Now().UTC().Unix()
	result, err := tx.ExecContext(ctx, `INSERT INTO Comments
    (userId, pageId, postedTime, content)
    VALUES(?,?,?,?)
    `, userId, pageId, now, comment.Content)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	if parentId == nil {
		_, err = tx.ExecContext(ctx, "INSERT INTO Relations(childId, depth) VALUES (?, 0)", id)
	} else {
		var parentDepth int64
		tx.QueryRowContext(ctx, "SELECT depth FROM Relations WHERE childId = ?", *parentId).Scan(&parentDepth)
		_, err = tx.ExecContext(ctx, "INSERT INTO Relations(parentId, childId, depth) (?, ?, ?)", *parentId, id, parentDepth+1)
	}
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	tx.Commit()
	return int(id), nil
}

func (p PennyDB) HideComment(ctx context.Context, commentId int64) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	now := time.Now().UTC().Unix()
	if _, err = tx.ExecContext(ctx, "UPDATE Comments SET hiddenTime = ? WHERE id = ?", now, commentId); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (p PennyDB) DeletComment(ctx context.Context, commentId int64) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	now := time.Now().UTC().Unix()
	_, err = tx.ExecContext(ctx, "UPDATE Comments SET deletedTime = ?, content = ? WHERE id = ?", now, "", commentId)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
