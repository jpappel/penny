package data

import (
	"context"
	"database/sql"
	"time"
)

func (p PennyDB) PostComment(ctx context.Context, page string, user string, comment string, parentId *int64) (int, error) {
	tx, err := p.Db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	var userId int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM Users WHERE email = ?", user).Scan(&userId)
	if err == sql.ErrNoRows {
		tx.Rollback()
        // TODO: gracefully handle no existing user
		return -1, err
	} else if err != nil {
		tx.Rollback()
		panic(err)
	}

	var pageId int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM Pages WHERE url = ?", page).Scan(&pageId)
	if err == sql.ErrNoRows {
        // TODO: gracefully handle no existing page
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
    `, userId, pageId, now, comment)
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
	tx, err := p.Db.BeginTx(ctx, nil)
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

func (p PennyDB) DeleteComment(ctx context.Context, commentId int64) error {
	tx, err := p.Db.BeginTx(ctx, nil)
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
