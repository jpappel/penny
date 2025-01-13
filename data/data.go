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

func initRelations(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS Relations(
        parentId INTEGER,
        childId INTEGER UNIQUE,
        depth INTEGER NOT NULL,
        FOREIGN KEY(parentId) REFERENCES Comments(id),
        FOREIGN KEY(childId) REFERENCES Comments(id)
    )`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE VIEW IF NOT EXISTS Parents AS
    WITH RECURSIVE parents(startId, parentId, childId) AS (
         SELECT childId, parentId, childId
         FROM Relations
         UNION ALL
         SELECT parents.startId, Relations.parentId, Relations.childId
         FROM Relations
         JOIN parents ON Relations.childId = parents.parentId
         )
    SELECT startId AS rootId, childId as commentId FROM parents WHERE parentId IS NULL
    `)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE VIEW IF NOT EXISTS Descendants AS
    WITH RECURSIVE descendants(startId, childId) AS (
        SELECT childId, childId
        FROM Relations
        UNION ALL
        SELECT descendants.startId, Relations.childId
        FROM Relations
        JOIN descendants ON Relations.parentId = descendants.childId
    )
    SELECT
        Comments.id AS commentId,
        COALESCE(descendants.descendantCount, 0) AS descendants,
        COALESCE(children.childCount, 0) AS children
    FROM Comments
    LEFT JOIN (
        SELECT startId AS commentId, COUNT(*) - 1 AS descendantCount
        FROM descendants
        GROUP BY startId
    ) descendants ON Comments.id = descendants.commentId
    LEFT JOIN (
        SELECT parentId AS commentId, COUNT(childId) AS childCount
        FROM Relations
        GROUP BY parentId
    ) children ON Comments.id = children.commentId`)
	if err != nil {
		panic(err)
	}
}

func InitDB(db *sql.DB) {
	initUsers(db)
	initPages(db)
	initComments(db)
	initRelations(db)
}

func New(connStr string) PennyDB {
	db := NewConn(connStr)
	InitDB(db)

	return PennyDB{db}
}
