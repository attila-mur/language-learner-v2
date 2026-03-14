package db

import (
	"database/sql"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	slog.Info("Database connected", "path", dbPath)

	if err = createTables(); err != nil {
		return err
	}

	return nil
}

func createTables() error {
	stmts := []string{
		// topics: no UNIQUE on name — enforced by partial index on seeded rows only
		`CREATE TABLE IF NOT EXISTS topics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    session_id TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES user_sessions(session_id)
)`,
		// seeded topics (session_id IS NULL) must have unique names
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_seeded_topic_name ON topics(name) WHERE session_id IS NULL`,
		`CREATE TABLE IF NOT EXISTS words (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    topic_id INTEGER NOT NULL,
    hungarian TEXT NOT NULL,
    english TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (topic_id) REFERENCES topics(id)
)`,
		`CREATE TABLE IF NOT EXISTS user_sessions (
    session_id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_accessed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    custom_words_count INTEGER DEFAULT 0
)`,
		`CREATE TABLE IF NOT EXISTS custom_words (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    topic_id INTEGER,
    hungarian TEXT NOT NULL,
    english TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES user_sessions(session_id),
    FOREIGN KEY (topic_id) REFERENCES topics(id)
)`,
		`CREATE TABLE IF NOT EXISTS scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    topic_id INTEGER,
    score INTEGER,
    total INTEGER,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES user_sessions(session_id),
    FOREIGN KEY (topic_id) REFERENCES topics(id)
)`,
	}

	for _, stmt := range stmts {
		if _, err := DB.Exec(stmt); err != nil {
			return err
		}
	}

	if err := migrateTopics(); err != nil {
		return err
	}

	slog.Info("Database tables created/verified")
	return nil
}

// migrateTopics handles upgrading existing databases that have the old topics schema
// (UNIQUE name, no session_id column) to the new schema.
func migrateTopics() error {
	// Check if session_id column already exists
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('topics') WHERE name='session_id'`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // already migrated
	}

	slog.Info("Migrating topics table to support custom topics...")

	stmts := []string{
		`CREATE TABLE topics_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    session_id TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES user_sessions(session_id)
)`,
		`INSERT INTO topics_new (id, name, description, created_at) SELECT id, name, description, created_at FROM topics`,
		`DROP TABLE topics`,
		`ALTER TABLE topics_new RENAME TO topics`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_seeded_topic_name ON topics(name) WHERE session_id IS NULL`,
	}

	for _, stmt := range stmts {
		if _, err := DB.Exec(stmt); err != nil {
			return err
		}
	}

	slog.Info("Topics table migration complete")
	return nil
}
