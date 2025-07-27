package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database connection
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	// Open database connection with foreign keys enabled
	connectionString := fmt.Sprintf("%s?_foreign_keys=on", dbPath)
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %v", err)
	}

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// DB returns the underlying database connection
func (d *Database) DB() *sql.DB {
	return d.db
}

// InitTables initializes all database tables
func (d *Database) InitTables() error {
	// Users table
	if err := d.createUsersTable(); err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// Sessions table
	if err := d.createSessionsTable(); err != nil {
		return fmt.Errorf("failed to create sessions table: %v", err)
	}

	return nil
}

func (d *Database) createUsersTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		session_limit INTEGER NOT NULL DEFAULT 5,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		created_at INTEGER NOT NULL,
		updated_at INTEGER
	)`

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createSessionsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS session_metadata (
		id TEXT PRIMARY KEY,
		phone TEXT NOT NULL,
		actual_phone TEXT,
		name TEXT,
		position INTEGER DEFAULT 0,
		webhook_url TEXT,
		created_at INTEGER NOT NULL
	)`

	_, err := d.db.Exec(query)
	return err
}