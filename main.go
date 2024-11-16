// hello world in go
package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open the database connection
	db, err := sql.Open("sqlite3", "./points.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the users table if it doesn't exist
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
        "user" TEXT NOT NULL PRIMARY KEY,
        "points" INTEGER NOT NULL
    );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Example usage
	addUser(db, "john_doe")
	addPoints(db, "john_doe", 5)
}

// addUser adds a new user to the database with 0 points
func addUser(db *sql.DB, username string) {
	insertUserSQL := `INSERT INTO users (user, points) VALUES (?, 0)`
	_, err := db.Exec(insertUserSQL, username)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User %s added successfully\n", username)
}

// addPoints increments the points for a given user
func addPoints(db *sql.DB, username string, points int) {
	updatePointsSQL := `UPDATE users SET points = points + ? WHERE user = ?`
	_, err := db.Exec(updatePointsSQL, points, username)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Added %d points to user %s\n", points, username)
}
