package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB
var get *sql.Stmt
var set *sql.Stmt

func openDb() {
	var err error
	database, err = sql.Open("sqlite3", appdir+"data.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = database.Exec("PRAGMA synchronous = OFF;")
	if err != nil {
		log.Fatal(err)
	}
	_, err = database.Exec("PRAGMA journal_mode = OFF;")
	if err != nil {
		log.Fatal(err)
	}
	_, err = database.Exec("CREATE TABLE IF NOT EXISTS Data (field STRING NOT NULL PRIMARY KEY, value text);")
	if err != nil {
		log.Fatal(err)
	}
	get, err = database.Prepare("SELECT field, value FROM data;")
	if err != nil {
		log.Fatal(err)
	}
	set, err = database.Prepare("INSERT OR REPLACE INTO Data (field, value) VALUES (?,?);")
	if err != nil {
		log.Fatal(err)
	}
}

func closeDb() {
	get.Close()
	set.Close()
	_, err := database.Exec("PRAGMA optimize;")
	if err != nil {
		log.Println(err)
	}
	database.Close()
}

func getData() {
	rows, err := get.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var field string
		var value string
		err = rows.Scan(&field, &value)
		if err != nil {
			log.Fatal(err)
		}
		data[field] = value
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func update(tx *sql.Tx, field string, value string) error {
	_, err := tx.Stmt(set).Exec(field, value)
	return err
}
