package dao

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func init() {
	var err error
	// https://briandouglas.ie/sqlite-defaults/
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
}

func MakeTables() {
	db.Exec(`
        CREATE TABLE IF NOT EXISTS polls (
            messageid TEXT NOT NULL,
            channelid TEXT NOT NULL,
            creatorid TEXT NOT NULL,
            points INTEGER NOT NULL,
            reason TEXT,
            expiry TIMESTAMP NOT NULL,
            passed BOOLEAN,
            PRIMARY KEY (messageid, channelid)
        );

        CREATE TABLE IF NOT EXISTS votes (
            messageid TEXT NOT NULL,
            channelid TEXT NOT NULL,
            voterid TEXT NOT NULL,
            FOREIGN KEY (messageid, channelid) REFERENCES polls(messageid, channelid)
        );

        CREATE TABLE IF NOT EXISTS gainers (
            messageid TEXT NOT NULL,
            channelid TEXT NOT NULL,
            gainerid TEXT NOT NULL,
            FOREIGN KEY (messageid, channelid) REFERENCES polls(messageid, channelid)
        );

        CREATE TABLE IF NOT EXISTS yearly (
            userid TEXT NOT NULL,
            year INTEGER NOT NULL,
            points INTEGER NOT NULL,
            PRIMARY KEY (userid, year)
        );
    `)
}
