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

	// CRM tables
	if err := d.createContactGroupsTable(); err != nil {
		return fmt.Errorf("failed to create contact_groups table: %v", err)
	}
	
	if err := d.createContactsTable(); err != nil {
		return fmt.Errorf("failed to create contacts table: %v", err)
	}

	if err := d.createMessageTemplatesTable(); err != nil {
		return fmt.Errorf("failed to create message_templates table: %v", err)
	}

	if err := d.createCampaignsTable(); err != nil {
		return fmt.Errorf("failed to create campaigns table: %v", err)
	}

	if err := d.createCampaignMessagesTable(); err != nil {
		return fmt.Errorf("failed to create campaign_messages table: %v", err)
	}

	if err := d.createAutoRepliesTable(); err != nil {
		return fmt.Errorf("failed to create auto_replies table: %v", err)
	}

	if err := d.createAutoReplyLogsTable(); err != nil {
		return fmt.Errorf("failed to create auto_reply_logs table: %v", err)
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

func (d *Database) createContactGroupsTable() error {
	var query string
	driver := d.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		query = `
		CREATE TABLE IF NOT EXISTS contact_groups (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			color VARCHAR(7),
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at BIGINT NOT NULL,
			updated_at BIGINT,
			INDEX idx_name (name),
			INDEX idx_is_active (is_active)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	} else {
		query = `
		CREATE TABLE IF NOT EXISTS contact_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			color TEXT,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL,
			updated_at INTEGER
		)`
		
		_, err := d.db.Exec(query)
		if err != nil {
			return err
		}
		
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_contact_groups_name ON contact_groups(name)",
			"CREATE INDEX IF NOT EXISTS idx_contact_groups_is_active ON contact_groups(is_active)",
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

func (d *Database) createContactsTable() error {
	var query string
	driver := d.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		query = `
		CREATE TABLE IF NOT EXISTS contacts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			phone VARCHAR(50) NOT NULL,
			email VARCHAR(255),
			company VARCHAR(255),
			position VARCHAR(255),
			group_id INT,
			tags JSON,
			notes TEXT,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			last_contact BIGINT,
			created_at BIGINT NOT NULL,
			updated_at BIGINT,
			FOREIGN KEY (group_id) REFERENCES contact_groups(id) ON DELETE SET NULL,
			INDEX idx_phone (phone),
			INDEX idx_name (name),
			INDEX idx_group_id (group_id),
			INDEX idx_is_active (is_active),
			INDEX idx_email (email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	} else {
		query = `
		CREATE TABLE IF NOT EXISTS contacts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			phone TEXT NOT NULL,
			email TEXT,
			company TEXT,
			position TEXT,
			group_id INTEGER,
			tags TEXT,
			notes TEXT,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			last_contact INTEGER,
			created_at INTEGER NOT NULL,
			updated_at INTEGER,
			FOREIGN KEY (group_id) REFERENCES contact_groups(id) ON DELETE SET NULL
		)`
		
		_, err := d.db.Exec(query)
		if err != nil {
			return err
		}
		
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_contacts_phone ON contacts(phone)",
			"CREATE INDEX IF NOT EXISTS idx_contacts_name ON contacts(name)",
			"CREATE INDEX IF NOT EXISTS idx_contacts_group_id ON contacts(group_id)",
			"CREATE INDEX IF NOT EXISTS idx_contacts_is_active ON contacts(is_active)",
			"CREATE INDEX IF NOT EXISTS idx_contacts_email ON contacts(email)",
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

func (d *Database) createMessageTemplatesTable() error {
	var query string
	driver := d.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		query = `
		CREATE TABLE IF NOT EXISTS message_templates (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'text',
			variables JSON,
			media_url TEXT,
			media_type VARCHAR(50),
			category VARCHAR(100),
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			usage_count INT NOT NULL DEFAULT 0,
			created_at BIGINT NOT NULL,
			updated_at BIGINT,
			INDEX idx_name (name),
			INDEX idx_type (type),
			INDEX idx_category (category),
			INDEX idx_is_active (is_active)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	} else {
		query = `
		CREATE TABLE IF NOT EXISTS message_templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			content TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'text',
			variables TEXT,
			media_url TEXT,
			media_type TEXT,
			category TEXT,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			usage_count INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER
		)`
		
		_, err := d.db.Exec(query)
		if err != nil {
			return err
		}
		
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_message_templates_name ON message_templates(name)",
			"CREATE INDEX IF NOT EXISTS idx_message_templates_type ON message_templates(type)",
			"CREATE INDEX IF NOT EXISTS idx_message_templates_category ON message_templates(category)",
			"CREATE INDEX IF NOT EXISTS idx_message_templates_is_active ON message_templates(is_active)",
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

func (d *Database) createCampaignsTable() error {
	var query string
	driver := d.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		query = `
		CREATE TABLE IF NOT EXISTS campaigns (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			template_id INT NOT NULL,
			group_id INT,
			contact_ids JSON,
			session_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'draft',
			delay_between INT NOT NULL DEFAULT 1,
			random_delay BOOLEAN NOT NULL DEFAULT FALSE,
			scheduled_at BIGINT,
			started_at BIGINT,
			completed_at BIGINT,
			total_contacts INT NOT NULL DEFAULT 0,
			sent_count INT NOT NULL DEFAULT 0,
			failed_count INT NOT NULL DEFAULT 0,
			pending_count INT NOT NULL DEFAULT 0,
			variables JSON,
			created_at BIGINT NOT NULL,
			updated_at BIGINT,
			FOREIGN KEY (template_id) REFERENCES message_templates(id) ON DELETE CASCADE,
			FOREIGN KEY (group_id) REFERENCES contact_groups(id) ON DELETE SET NULL,
			INDEX idx_name (name),
			INDEX idx_status (status),
			INDEX idx_session_id (session_id),
			INDEX idx_template_id (template_id),
			INDEX idx_group_id (group_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	} else {
		query = `
		CREATE TABLE IF NOT EXISTS campaigns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			template_id INTEGER NOT NULL,
			group_id INTEGER,
			contact_ids TEXT,
			session_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'draft',
			delay_between INTEGER NOT NULL DEFAULT 1,
			random_delay BOOLEAN NOT NULL DEFAULT 0,
			scheduled_at INTEGER,
			started_at INTEGER,
			completed_at INTEGER,
			total_contacts INTEGER NOT NULL DEFAULT 0,
			sent_count INTEGER NOT NULL DEFAULT 0,
			failed_count INTEGER NOT NULL DEFAULT 0,
			pending_count INTEGER NOT NULL DEFAULT 0,
			variables TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER,
			FOREIGN KEY (template_id) REFERENCES message_templates(id) ON DELETE CASCADE,
			FOREIGN KEY (group_id) REFERENCES contact_groups(id) ON DELETE SET NULL
		)`
		
		_, err := d.db.Exec(query)
		if err != nil {
			return err
		}
		
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_campaigns_name ON campaigns(name)",
			"CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status)",
			"CREATE INDEX IF NOT EXISTS idx_campaigns_session_id ON campaigns(session_id)",
			"CREATE INDEX IF NOT EXISTS idx_campaigns_template_id ON campaigns(template_id)",
			"CREATE INDEX IF NOT EXISTS idx_campaigns_group_id ON campaigns(group_id)",
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

func (d *Database) createCampaignMessagesTable() error {
	var query string
	driver := d.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		query = `
		CREATE TABLE IF NOT EXISTS campaign_messages (
			id INT AUTO_INCREMENT PRIMARY KEY,
			campaign_id INT NOT NULL,
			contact_id INT NOT NULL,
			content TEXT NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			error_msg TEXT,
			message_id VARCHAR(255),
			sent_at BIGINT,
			created_at BIGINT NOT NULL,
			FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE,
			FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE,
			INDEX idx_campaign_id (campaign_id),
			INDEX idx_contact_id (contact_id),
			INDEX idx_status (status),
			INDEX idx_sent_at (sent_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	} else {
		query = `
		CREATE TABLE IF NOT EXISTS campaign_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			campaign_id INTEGER NOT NULL,
			contact_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			error_msg TEXT,
			message_id TEXT,
			sent_at INTEGER,
			created_at INTEGER NOT NULL,
			FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE,
			FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
		)`
		
		_, err := d.db.Exec(query)
		if err != nil {
			return err
		}
		
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_campaign_messages_campaign_id ON campaign_messages(campaign_id)",
			"CREATE INDEX IF NOT EXISTS idx_campaign_messages_contact_id ON campaign_messages(contact_id)",
			"CREATE INDEX IF NOT EXISTS idx_campaign_messages_status ON campaign_messages(status)",
			"CREATE INDEX IF NOT EXISTS idx_campaign_messages_sent_at ON campaign_messages(sent_at)",
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

func (d *Database) createAutoRepliesTable() error {
	var query string
	driver := d.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		query = `
		CREATE TABLE IF NOT EXISTS auto_replies (
			id INT AUTO_INCREMENT PRIMARY KEY,
			session_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			trigger_type VARCHAR(50) NOT NULL,
			keywords JSON,
			response TEXT NOT NULL,
			media_url TEXT,
			media_type VARCHAR(50),
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			priority INT NOT NULL DEFAULT 0,
			delay_min INT NOT NULL DEFAULT 0,
			delay_max INT NOT NULL DEFAULT 0,
			max_replies INT NOT NULL DEFAULT 0,
			time_start VARCHAR(5),
			time_end VARCHAR(5),
			conditions JSON,
			usage_count INT NOT NULL DEFAULT 0,
			created_at BIGINT NOT NULL,
			updated_at BIGINT,
			INDEX idx_session_id (session_id),
			INDEX idx_trigger_type (trigger_type),
			INDEX idx_is_active (is_active),
			INDEX idx_priority (priority)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	} else {
		query = `
		CREATE TABLE IF NOT EXISTS auto_replies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			name TEXT NOT NULL,
			trigger_type TEXT NOT NULL,
			keywords TEXT,
			response TEXT NOT NULL,
			media_url TEXT,
			media_type TEXT,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			priority INTEGER NOT NULL DEFAULT 0,
			delay_min INTEGER NOT NULL DEFAULT 0,
			delay_max INTEGER NOT NULL DEFAULT 0,
			max_replies INTEGER NOT NULL DEFAULT 0,
			time_start TEXT,
			time_end TEXT,
			conditions TEXT,
			usage_count INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER
		)`
		
		_, err := d.db.Exec(query)
		if err != nil {
			return err
		}
		
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_auto_replies_session_id ON auto_replies(session_id)",
			"CREATE INDEX IF NOT EXISTS idx_auto_replies_trigger_type ON auto_replies(trigger_type)",
			"CREATE INDEX IF NOT EXISTS idx_auto_replies_is_active ON auto_replies(is_active)",
			"CREATE INDEX IF NOT EXISTS idx_auto_replies_priority ON auto_replies(priority)",
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

func (d *Database) createAutoReplyLogsTable() error {
	var query string
	driver := d.db.Driver()
	driverName := fmt.Sprintf("%T", driver)
	
	if contains(driverName, "mysql") {
		query = `
		CREATE TABLE IF NOT EXISTS auto_reply_logs (
			id INT AUTO_INCREMENT PRIMARY KEY,
			auto_reply_id INT NOT NULL,
			session_id VARCHAR(255) NOT NULL,
			contact_phone VARCHAR(50) NOT NULL,
			trigger_msg TEXT NOT NULL,
			response TEXT NOT NULL,
			success BOOLEAN NOT NULL DEFAULT TRUE,
			error_msg TEXT,
			created_at BIGINT NOT NULL,
			FOREIGN KEY (auto_reply_id) REFERENCES auto_replies(id) ON DELETE CASCADE,
			INDEX idx_auto_reply_id (auto_reply_id),
			INDEX idx_session_id (session_id),
			INDEX idx_contact_phone (contact_phone),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	} else {
		query = `
		CREATE TABLE IF NOT EXISTS auto_reply_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			auto_reply_id INTEGER NOT NULL,
			session_id TEXT NOT NULL,
			contact_phone TEXT NOT NULL,
			trigger_msg TEXT NOT NULL,
			response TEXT NOT NULL,
			success BOOLEAN NOT NULL DEFAULT 1,
			error_msg TEXT,
			created_at INTEGER NOT NULL,
			FOREIGN KEY (auto_reply_id) REFERENCES auto_replies(id) ON DELETE CASCADE
		)`
		
		_, err := d.db.Exec(query)
		if err != nil {
			return err
		}
		
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_auto_reply_logs_auto_reply_id ON auto_reply_logs(auto_reply_id)",
			"CREATE INDEX IF NOT EXISTS idx_auto_reply_logs_session_id ON auto_reply_logs(session_id)",
			"CREATE INDEX IF NOT EXISTS idx_auto_reply_logs_contact_phone ON auto_reply_logs(contact_phone)",
			"CREATE INDEX IF NOT EXISTS idx_auto_reply_logs_created_at ON auto_reply_logs(created_at)",
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