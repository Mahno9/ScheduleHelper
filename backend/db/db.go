package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	return createTables()
}

func createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		color TEXT NOT NULL,
		emoji TEXT NOT NULL,
		theme TEXT DEFAULT 'system',
		timezone TEXT DEFAULT 'auto',
		password_hash TEXT,
		google_calendar_token TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS slots (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		comment TEXT,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		icon TEXT,
		color TEXT,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS event_participants (
		event_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		PRIMARY KEY (event_id, user_id),
		FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	);
	`
	_, err := DB.Exec(query)
	if err != nil {
		log.Printf("Error creating tables: %v", err)
	}
	return err
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}