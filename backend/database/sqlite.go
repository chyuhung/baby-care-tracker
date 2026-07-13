package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return err
	}

	DB.SetMaxOpenConns(1) // SQLite 单线程写入

	if err = DB.Ping(); err != nil {
		return err
	}

	if err = createTables(); err != nil {
		return err
	}

	log.Printf("✅ SQLite 数据库已初始化: %s", dbPath)
	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS families (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		invite_code TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		family_id INTEGER REFERENCES families(id),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS babies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		birth_date DATE NOT NULL,
		gender TEXT DEFAULT '',
		avatar_color TEXT DEFAULT '#6C63FF',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS feeding_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		baby_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		duration_minutes INTEGER DEFAULT 0,
		amount_ml INTEGER DEFAULT 0,
		side TEXT DEFAULT '',
		brand TEXT DEFAULT '',
		note TEXT DEFAULT '',
		occurred_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (baby_id) REFERENCES babies(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS diaper_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		baby_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		note TEXT DEFAULT '',
		occurred_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (baby_id) REFERENCES babies(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_babies_user ON babies(user_id);
	CREATE INDEX IF NOT EXISTS idx_feeding_baby ON feeding_records(baby_id);
	CREATE INDEX IF NOT EXISTS idx_feeding_occurred ON feeding_records(occurred_at);
	CREATE INDEX IF NOT EXISTS idx_diaper_baby ON diaper_records(baby_id);
	CREATE INDEX IF NOT EXISTS idx_diaper_occurred ON diaper_records(occurred_at);
	`

	_, err := DB.Exec(schema)
	if err != nil {
		return err
	}

	// Migration: add family_id column if it doesn't exist (for existing databases)
	DB.Exec("ALTER TABLE users ADD COLUMN family_id INTEGER REFERENCES families(id)")

	return nil
}
