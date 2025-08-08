package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Database represents the database connection
type Database struct {
	db *sql.DB
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// NewDatabase creates a new database connection
func NewDatabase(config DatabaseConfig) (*Database, error) {
	// MySQL connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		config.User, config.Password, config.Host, config.Port, config.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL database: %v", err)
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

	// Messages table
	if err := d.createMessagesTable(); err != nil {
		return fmt.Errorf("failed to create messages table: %v", err)
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
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			api_key VARCHAR(64) UNIQUE NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			session_limit INT NOT NULL DEFAULT 5,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at BIGINT NOT NULL,
			updated_at BIGINT
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err := d.db.Exec(query)
	if err != nil {
		return err
	}

	// Add API key column if it doesn't exist (migration for existing databases)
	return d.migrateUsersTable()
}

// migrateUsersTable adds missing columns to existing users table
func (d *Database) migrateUsersTable() error {
	// Check if api_key column exists
	query := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'users' 
		AND COLUMN_NAME = 'api_key'`

	var count int
	err := d.db.QueryRow(query).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = d.db.Exec("ALTER TABLE users ADD COLUMN api_key VARCHAR(64) UNIQUE NULL")
		return err
	}

	return nil
}

func (d *Database) createSessionsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS session_metadata (
			id VARCHAR(255) PRIMARY KEY,
			phone VARCHAR(50) NOT NULL,
			actual_phone VARCHAR(50),
			name VARCHAR(255),
			position INT DEFAULT 0,
			webhook_url TEXT,
			auto_reply_text TEXT,
			proxy_enabled BOOLEAN DEFAULT FALSE,
			proxy_type VARCHAR(10) DEFAULT '',
			proxy_host VARCHAR(255) DEFAULT '',
			proxy_port INT DEFAULT 0,
			proxy_username VARCHAR(255) DEFAULT '',
			proxy_password VARCHAR(255) DEFAULT '',
			enabled BOOLEAN DEFAULT TRUE,
			user_id INT NOT NULL DEFAULT 1,
			created_at BIGINT NOT NULL,
			INDEX idx_phone (phone),
			INDEX idx_user_id (user_id),
			INDEX idx_created_at (created_at),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err := d.db.Exec(query)
	if err != nil {
		return err
	}

	// Migrate existing session_metadata table to add enabled column if missing
	return d.migrateSessionsTable()
}

// migrateSessionsTable adds missing columns to existing session_metadata table
func (d *Database) migrateSessionsTable() error {
	// Check if enabled column exists
	query := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'session_metadata' 
		AND COLUMN_NAME = 'enabled'`
	
	var count int
	err := d.db.QueryRow(query).Scan(&count)
	if err != nil {
		return err
	}
	
	if count == 0 {
		_, err = d.db.Exec("ALTER TABLE session_metadata ADD COLUMN enabled BOOLEAN DEFAULT TRUE")
		if err != nil {
			return err
		}
		// Update existing records to have enabled = TRUE
		_, err = d.db.Exec("UPDATE session_metadata SET enabled = TRUE WHERE enabled IS NULL")
		return err
	}
	
	return nil
}

func (d *Database) createMessagesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS messages (
			id INT AUTO_INCREMENT PRIMARY KEY,
			session_id VARCHAR(255) NOT NULL,
			message_id VARCHAR(255) UNIQUE,
			sender_jid VARCHAR(100),
			recipient_jid VARCHAR(100),
			message_type VARCHAR(50) NOT NULL DEFAULT 'text',
			content TEXT,
			media_url TEXT,
			direction VARCHAR(20) NOT NULL,
			status VARCHAR(50),
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_session_id (session_id),
			INDEX idx_sender_jid (sender_jid),
			INDEX idx_recipient_jid (recipient_jid),
			INDEX idx_direction (direction),
			INDEX idx_status (status),
			INDEX idx_created_at (created_at),
			FOREIGN KEY (session_id) REFERENCES session_metadata(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createLogsTable() error {
	query := `
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

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createContactGroupsTable() error {
	query := `
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

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createContactsTable() error {
	query := `
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

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createMessageTemplatesTable() error {
	query := `
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

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createCampaignsTable() error {
	query := `
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

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createCampaignMessagesTable() error {
	query := `
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

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createAutoRepliesTable() error {
	query := `
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

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createAutoReplyLogsTable() error {
	query := `
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

	_, err := d.db.Exec(query)
	return err
}