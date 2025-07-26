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

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Update admin user
	_, err = db.Exec("UPDATE users SET password_hash = ? WHERE username = 'admin'", string(hashedPassword))
	if err != nil {
		log.Fatal("Failed to update admin password:", err)
	}

	fmt.Println("Admin password updated successfully")
	
	// Test the password
	var storedHash string
	err = db.QueryRow("SELECT password_hash FROM users WHERE username = 'admin'").Scan(&storedHash)
	if err != nil {
		log.Fatal("Failed to retrieve password:", err)
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("admin123"))
	if err != nil {
		fmt.Println("Password verification failed:", err)
	} else {
		fmt.Println("Password verification succeeded")
	}
}