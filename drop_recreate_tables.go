package main

import (
	"database/sql"
	"log"
	
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Connect to MySQL
	dsn := "root:robioki@tcp(127.0.0.1:3306)/whatsapGo"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("Dropping existing tables...")
	
	// Disable foreign key checks
	_, err = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	if err != nil {
		log.Fatal("Failed to disable foreign key checks:", err)
	}
	
	// Drop all tables
	tables := []string{"user_auth_sessions", "session_metadata", "users", "whatsmeow_device", "whatsmeow_identity_keys", "whatsmeow_prekeys", "whatsmeow_sessions", "whatsmeow_sender_keys", "whatsmeow_app_state_sync_keys", "whatsmeow_app_state_version", "whatsmeow_app_state_mutation_macs", "whatsmeow_contacts", "whatsmeow_chat_settings", "whatsmeow_message_secrets"}
	
	for _, table := range tables {
		_, err = db.Exec("DROP TABLE IF EXISTS " + table)
		if err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		} else {
			log.Printf("Dropped table %s", table)
		}
	}
	
	// Re-enable foreign key checks
	_, err = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	if err != nil {
		log.Fatal("Failed to re-enable foreign key checks:", err)
	}

	log.Println("Creating fresh tables...")

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at BIGINT NOT NULL,
			INDEX idx_username (username),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}
	log.Println("Created users table")

	// Create session_metadata table
	_, err = db.Exec(`
		CREATE TABLE session_metadata (
			id VARCHAR(255) PRIMARY KEY,
			phone VARCHAR(255) NOT NULL,
			actual_phone VARCHAR(255),
			name VARCHAR(255),
			created_at BIGINT NOT NULL,
			INDEX idx_phone (phone),
			INDEX idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	if err != nil {
		log.Fatal("Failed to create session_metadata table:", err)
	}
	log.Println("Created session_metadata table")

	// Create admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (username, password_hash, created_at) 
		VALUES (?, ?, UNIX_TIMESTAMP())
	`, "admin", string(hashedPassword))
	if err != nil {
		log.Fatal("Failed to insert admin user:", err)
	}
	log.Println("Created admin user")

	log.Println("Database tables recreated successfully!")
}