package data

import (
	"database/sql"

	_ "github.com/tursodatabase/go-libsql"
)

var DB *sql.DB

func NewConn(connStr string) *sql.DB {
	db, err := sql.Open("libsql", connStr)
	if err != nil {
		panic(err)
	}

	return db
}

func initUsers(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS Users(
        id INTEGER PRIMARY KEY,
        email TEXT,
        provider TEXT NOT NULL,
        name TEXT,
        UNIQUE(email, provider)
    )`)
	if err != nil {
		panic(err)
	}
}

func initPages(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS Pages(
        id INTEGER PRIMARY KEY,
        url TEXT UNIQUE NOT NULL,
        commentsOpenTime INTEGER
    )`)
	if err != nil {
		panic(err)
	}
}

func initComments(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS Comments(
        id INTEGER PRIMARY KEY,
        userId INTEGER NOT NULL,
        pageId INTEGER NOT NULL,
        hiddenTime INTEGER,
        deletedTime INTEGER,
        postedTime INTEGER NOT NULL,
        content TEXT NOT NULL,
        FOREIGN KEY(userId) REFERENCES Users(id),
        FOREIGN KEY(pageId) REFERENCES Pages(id)
    )`)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_users ON Comments(userId)")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_postedTime ON Comments(postedTime)")
	if err != nil {
		panic(err)
	}
}

func initReplies(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS Replies(
        id INTEGER PRIMARY KEY,
        parentId INTEGER NOT NULL,
        childId INTEGER NOT NULL,
        FOREIGN KEY(parentId) REFERENCES Comments(id),
        FOREIGN KEY(childId) REFERENCES Comments(id)
    )`)
	if err != nil {
		panic(err)
	}
}

func InitDB(db *sql.DB) {
	initUsers(db)
	initPages(db)
	initComments(db)
	initReplies(db)
}

func New(connStr string) PennyDB {
	db := NewConn(connStr)
	InitDB(db)

	return PennyDB{db}
}
