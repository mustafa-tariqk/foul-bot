package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite", "file:foulbot.sqlite?"+
		"_journal_mode=WAL&"+
		"_synchronous=NORMAL&"+
		"_busy_timeout=5000&"+
		"_cache_size=-20000&"+
		"_foreign_keys=ON&"+
		"_auto_vacuum=INCREMENTAL&"+
		"_temp_store=MEMORY&"+
		"_mmap_size=2147483648&"+
		"_page_size=8192")
	if err != nil {
		log.Fatal(err)
	}

	makeTables()
}

func makeTables() {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS polls (
            messageId TEXT PRIMARY KEY,
            channelId TEXT,
            points INTEGER,
            reason TEXT,
            expiry TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS votes (
            messageId TEXT,
            channelId TEXT,
			userId TEXT,
            upvote BOOLEAN,
            FOREIGN KEY (messageId) REFERENCES polls(messageId),
            FOREIGN KEY (channelId) REFERENCES polls(channelId)
        );
        CREATE TABLE IF NOT EXISTS accused (
            messageId TEXT PRIMARY KEY,
            channelId TEXT,
            userId TEXT
        );
    `)
	if err != nil {
		log.Fatal(err)
	}
}

func createPoll(messageId string, channelId string, points int, reason string, expiry string, userIds []string) {
	_, err := db.Exec(`
		INSERT INTO polls (messageId, channelId, points, reason, expiry)
		VALUES (?, ?, ?, ?, ?)
	`, messageId, channelId, points, reason, expiry)

	if err != nil {
		log.Fatal(err)
	}

	for _, userId := range userIds {
		_, err = db.Exec(`
			INSERT INTO accused (messageId, channelId, userId)
			VALUES (?, ?, ?)
		`, messageId, channelId, userId)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func concludePoll(messageId string, channelId string, usersFor []string, usersAgainst []string) {
	query := `
	INSERT INTO votes (messageId, channelId, userId, upvote)
	VALUES (?, ?, ?, ?)
`

	for _, userId := range usersFor {
		_, err := db.Exec(query, messageId, channelId, userId, true)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, userId := range usersAgainst {
		_, err := db.Exec(query, messageId, channelId, userId, false)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func main() {
	// while true
	reader := bufio.NewReader(os.Stdin)
	for {
		// get user input
		fmt.Print("Enter command: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input)

		// if user input is "exit"
		if input == "exit" {
			break
		} else if input == "create" {
			createPoll("123", "456", 1, "reason", "expiry", []string{"789"})
		}

		// handle other commands
		fmt.Println("You entered:", input)
	}
}
