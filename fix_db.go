package main

import (
	"database/sql"
	"fmt"
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

	// Check current table structure
	fmt.Println("=== Current users table structure ===")
	rows, err := db.Query("DESCRIBE users")
	if err != nil {
		log.Fatal("Failed to describe users table:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var field, typ, null, key, dflt, extra sql.NullString
		err := rows.Scan(&field, &typ, &null, &key, &dflt, &extra)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}
		fmt.Printf("Field: %s, Type: %s, Null: %s, Key: %s, Default: %s, Extra: %s\n", 
			field.String, typ.String, null.String, key.String, dflt.String, extra.String)
	}

	// Check if password column exists
	var columnExists bool
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='whatsapGo' AND table_name='users' AND column_name='password'").Scan(&columnExists)
	if err != nil {
		log.Fatal("Failed to check password column:", err)
	}

	if !columnExists {
		fmt.Println("\n=== Adding password column ===")
		_, err = db.Exec("ALTER TABLE users ADD COLUMN password VARCHAR(255) NOT NULL DEFAULT ''")
		if err != nil {
			log.Fatal("Failed to add password column:", err)
		}
		fmt.Println("Password column added successfully")
	} else {
		fmt.Println("\n=== Password column already exists ===")
	}

	// Check if admin user exists and update password
	fmt.Println("\n=== Checking admin user ===")
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		log.Fatal("Failed to check admin user:", err)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	if count > 0 {
		// Update existing admin user
		_, err = db.Exec("UPDATE users SET password = ? WHERE username = 'admin'", string(hashedPassword))
		if err != nil {
			log.Fatal("Failed to update admin password:", err)
		}
		fmt.Println("Admin password updated successfully")
	} else {
		// Insert new admin user
		_, err = db.Exec("INSERT INTO users (username, password, created_at) VALUES ('admin', ?, UNIX_TIMESTAMP())", string(hashedPassword))
		if err != nil {
			log.Fatal("Failed to insert admin user:", err)
		}
		fmt.Println("Admin user created successfully")
	}

	fmt.Println("\n=== Database fix completed ===")
}