package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	var err error
	DB, err = sql.Open("sqlite3", "./logitwin.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		otp TEXT,
		otp_expiry TEXT
	);

	CREATE TABLE IF NOT EXISTS nodes (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		lat REAL NOT NULL,
		lng REAL NOT NULL,
		type TEXT NOT NULL,
		status TEXT NOT NULL,
		stock INTEGER DEFAULT 0,
		capacity INTEGER DEFAULT 1000
	);

	CREATE TABLE IF NOT EXISTS vehicles (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		status TEXT NOT NULL,
		load INTEGER DEFAULT 0,
		lat REAL NOT NULL,
		lng REAL NOT NULL,
		origin_id TEXT NOT NULL,
		dest_id TEXT NOT NULL,
		progress REAL DEFAULT 0,
		eta TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		time TEXT NOT NULL,
		type TEXT NOT NULL,
		message TEXT NOT NULL,
		severity TEXT NOT NULL
	);
	`

	_, err = DB.Exec(createTables)
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	fmt.Println("✅ Database initialized")
}