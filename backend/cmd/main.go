package main

import (
	"database/sql"
	"fmt"
	"log"

	"real-time/backend/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Setup in-memory database
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := repository.MigrateDB(db); err != nil {
		log.Fatal("Migration failed:", err)
	}
	fmt.Println("Database migrated successfully!")

	// Test operations
	testUserOperations(db)
}

func testUserOperations(db *sql.DB) {
	// Insert test user
	res, err := db.Exec(`
		INSERT INTO users 
		(uuid, nickname, email, password_hash, first_name, last_name, age, gender) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"test-uuid-123", "testuser1", "test1@example.com",
		"$2a$10$hashedpassword", "John", "Doe", 30, "male")
	if err != nil {
		log.Fatal("Insert failed:", err)
	}

	// Get inserted ID
	id, _ := res.LastInsertId()
	fmt.Printf("Inserted user with ID %d\n", id)

	// Query test
	var nickname string
	err = db.QueryRow("SELECT nickname FROM users WHERE id = ?", id).Scan(&nickname)
	if err != nil {
		log.Fatal("Query failed:", err)
	}
	fmt.Println("Found user:", nickname)
}
