package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database connection
type Database struct {
	db *sql.DB
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string // "sqlite" or "mysql"
	Path     string // for SQLite
	Host     string // for MySQL
	Port     string // for MySQL
	User     string // for MySQL
	Password string // for MySQL
	Database string // for MySQL
}

// NewDatabase creates a new database connection
func NewDatabase(config DatabaseConfig) (*Database, error) {
	var db *sql.DB
	var err error

	if config.Type == "mysql" {
		// MySQL connection
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
			config.User, config.Password, config.Host, config.Port, config.Database)
		
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open MySQL database: %v", err)
		}
	} else {
		// SQLite connection (default)
		// Ensure directory exists
		dir := filepath.Dir(config.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %v", err)
		}

		// Open database connection with foreign keys enabled
		connectionString := fmt.Sprintf("%s?_foreign_keys=on", config.Path)
		db, err = sql.Open("sqlite3", connectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite database: %v", err)
		}

		// Enable foreign keys for SQLite
		if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
			return nil, fmt.Errorf("failed to enable foreign keys: %v", err)
		}
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
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

	// Logs table
	if err := d.createLogsTable(); err != nil {
		return fmt.Errorf("failed to create logs table: %v", err)
	}

	return nil
}

func (d *Database) createUsersTable() error {
	// Check if we're using MySQL or SQLite
	var query string
	
	// Get database driver name
	driver := d.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		query = `
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			session_limit INT NOT NULL DEFAULT 5,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at BIGINT NOT NULL,
			updated_at BIGINT
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	} else {
		query = `
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
	}

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createSessionsTable() error {
	// Check if we're using MySQL or SQLite
	var query string
	
	// Get database driver name
	driver := d.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		query = `
		CREATE TABLE IF NOT EXISTS session_metadata (
			id VARCHAR(255) PRIMARY KEY,
			phone VARCHAR(50) NOT NULL,
			actual_phone VARCHAR(50),
			name VARCHAR(255),
			position INT DEFAULT 0,
			webhook_url TEXT,
			created_at BIGINT NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	} else {
		query = `
		CREATE TABLE IF NOT EXISTS session_metadata (
			id TEXT PRIMARY KEY,
			phone TEXT NOT NULL,
			actual_phone TEXT,
			name TEXT,
			position INTEGER DEFAULT 0,
			webhook_url TEXT,
			created_at INTEGER NOT NULL
		)`
	}

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createLogsTable() error {
	// Check if we're using MySQL or SQLite
	var query string
	
	// Get database driver name
	driver := d.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		query = `
		CREATE TABLE IF NOT EXISTS logs (
			id INT AUTO_INCREMENT PRIMARY KEY,
			level VARCHAR(10) NOT NULL,
			message TEXT NOT NULL,
			component VARCHAR(100),
			session_id VARCHAR(255),
			user_id INT,
			metadata JSON,
			created_at BIGINT NOT NULL,
			INDEX idx_level (level),
			INDEX idx_component (component),
			INDEX idx_session_id (session_id),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	} else {
		query = `
		CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			component TEXT,
			session_id TEXT,
			user_id INTEGER,
			metadata TEXT,
			created_at INTEGER NOT NULL
		)`
		
		// Create indexes for SQLite
		_, err := d.db.Exec(query)
		if err != nil {
			return err
		}
		
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level)",
			"CREATE INDEX IF NOT EXISTS idx_logs_component ON logs(component)",
			"CREATE INDEX IF NOT EXISTS idx_logs_session_id ON logs(session_id)",
			"CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at)",
		}
		
		for _, indexQuery := range indexes {
			if _, err := d.db.Exec(indexQuery); err != nil {
				return err
			}
		}
		
		return nil
	}

	_, err := d.db.Exec(query)
	return err
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(hasPrefix(s, substr) || hasSuffix(s, substr) || indexString(s, substr) >= 0))
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}