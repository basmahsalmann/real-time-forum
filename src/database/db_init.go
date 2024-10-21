package forum

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var Db *sql.DB

// Initialize the database connection and create tables
func InitDB() {
	var err error
	Db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal(err)
	}
	createTables()
}

func createTables() {
	schema := `
	CREATE TABLE IF NOT EXISTS user (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nickname TEXT UNIQUE NOT NULL,
		age INTEGER,
		gender TEXT,
		first_name TEXT,
		last_name TEXT,
		password TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS post (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES user (id)
	);

	CREATE TABLE IF NOT EXISTS category (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		user_id INTEGER,
		FOREIGN KEY (user_id) REFERENCES user (id)
	);

	CREATE TABLE IF NOT EXISTS post_category (
		post_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		PRIMARY KEY (post_id, category_id),
		FOREIGN KEY (post_id) REFERENCES post (id),
		FOREIGN KEY (category_id) REFERENCES category (id)
	);

	CREATE TABLE IF NOT EXISTS comments (
		comment_id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		FOREIGN KEY (post_id) REFERENCES post (id),
		FOREIGN KEY (user_id) REFERENCES user (id)
	);

	CREATE TABLE IF NOT EXISTS post_likes_dislikes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		post_id INTEGER NOT NULL,
		type TEXT CHECK(type IN ('like', 'dislike')) NOT NULL,
		FOREIGN KEY (user_id) REFERENCES user (id),
		FOREIGN KEY (post_id) REFERENCES post (id)
	);

	CREATE TABLE IF NOT EXISTS comment_likes_dislikes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		comment_id INTEGER NOT NULL,
		type TEXT CHECK(type IN ('like', 'dislike')) NOT NULL,
		FOREIGN KEY (user_id) REFERENCES user (id),
		FOREIGN KEY (comment_id) REFERENCES comments (comment_id)
	);

	CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id)
	);
	`

	_, err := Db.Exec(schema)
	if err != nil {
		log.Fatal("Error creating tables: ", err)
	}
}
