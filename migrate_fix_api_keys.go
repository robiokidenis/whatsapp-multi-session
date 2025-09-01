package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Get database configuration
	dbType := os.Getenv("DATABASE_TYPE")
	if dbType == "" {
		dbType = "mysql" // default to MySQL for production
	}

	var db *sql.DB
	var err error

	if dbType == "mysql" {
		// MySQL connection
		host := os.Getenv("MYSQL_HOST")
		port := os.Getenv("MYSQL_PORT")
		user := os.Getenv("MYSQL_USER")
		password := os.Getenv("MYSQL_PASSWORD")
		database := os.Getenv("MYSQL_DATABASE")

		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "3306"
		}

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			user, password, host, port, database)

		log.Printf("Connecting to MySQL database: %s@%s:%s/%s\n", user, host, port, database)
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Fatalf("Failed to connect to MySQL: %v", err)
		}
	} else {
		log.Fatalf("Unsupported database type: %s", dbType)
	}

	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to database successfully")

	// Step 1: Check for empty API keys
	log.Println("\n📊 Checking for empty API keys...")
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE api_key = ''").Scan(&count)
	if err != nil {
		log.Printf("Warning: Failed to count empty API keys: %v", err)
	} else {
		log.Printf("Found %d users with empty API keys\n", count)
	}

	// Step 2: Show affected users
	if count > 0 {
		rows, err := db.Query("SELECT id, username FROM users WHERE api_key = ''")
		if err != nil {
			log.Printf("Warning: Failed to list affected users: %v", err)
		} else {
			defer rows.Close()
			log.Println("\n👥 Affected users:")
			for rows.Next() {
				var id int
				var username string
				if err := rows.Scan(&id, &username); err == nil {
					log.Printf("  - ID: %d, Username: %s", id, username)
				}
			}
		}
	}

	// Step 3: Fix empty API keys by setting them to NULL
	if count > 0 {
		log.Println("\n🔧 Fixing empty API keys by setting them to NULL...")
		result, err := db.Exec("UPDATE users SET api_key = NULL WHERE api_key = ''")
		if err != nil {
			log.Fatalf("Failed to update empty API keys: %v", err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			log.Printf("Warning: Could not get affected rows: %v", err)
		} else {
			log.Printf("✅ Updated %d rows successfully", affected)
		}
	} else {
		log.Println("\n✅ No empty API keys found. Database is already clean!")
	}

	// Step 4: Verify the fix
	log.Println("\n🔍 Verifying the fix...")
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE api_key = ''").Scan(&count)
	if err != nil {
		log.Printf("Warning: Failed to verify fix: %v", err)
	} else {
		if count == 0 {
			log.Println("✅ SUCCESS: No more empty API keys in the database!")
		} else {
			log.Printf("⚠️  WARNING: Still found %d empty API keys", count)
		}
	}

	// Step 5: Show final status
	log.Println("\n📊 Final API key status:")
	rows, err := db.Query(`
		SELECT 
			CASE 
				WHEN api_key IS NULL THEN 'NULL (valid)'
				WHEN api_key = '' THEN 'EMPTY (invalid)'
				ELSE 'SET'
			END as status,
			COUNT(*) as count
		FROM users
		GROUP BY status
	`)
	if err != nil {
		log.Printf("Warning: Failed to get final status: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			if err := rows.Scan(&status, &count); err == nil {
				log.Printf("  %s: %d users", status, count)
			}
		}
	}

	log.Println("\n✅ Migration completed successfully!")
	log.Println("You can now create new users without the duplicate key error.")
}