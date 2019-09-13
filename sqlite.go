package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB
var get *sql.Stmt
var set *sql.Stmt
var del *sql.Stmt
var buffer bytes.Buffer

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

	del, err = database.Prepare("DELETE FROM Data WHERE field = ?;")
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
	fmt.Println("\nGood bye!")
}

func getData() []byte {
	if buffer.Len() > 0 {
		return buffer.Bytes()
	}
	rows, err := get.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	buffer.WriteString("{")
	for rows.Next() {
		var field []byte
		var value []byte
		err = rows.Scan(&field, &value)
		if err != nil {
			log.Fatal(err)
		}
		buffer.WriteString(`"`)
		buffer.Write(field)
		buffer.WriteString(`":`)
		buffer.Write(value)
		buffer.WriteString(",")
	}
	if buffer.Len() > 1 {
		buffer.Truncate(buffer.Len() - 1)
	}
	buffer.WriteString("}")
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return buffer.Bytes()
}

func update(tx *sql.Tx, field string, value string) error {
	var err error
	if value == "null" {
		_, err = tx.Stmt(del).Exec(field)
	} else {
		_, err = tx.Stmt(set).Exec(field, value)
	}
	return err
}

func dropDb(tx *sql.Tx) error {
	_, err := database.Exec("DROP TABLE Data")
	_, err = database.Exec("CREATE TABLE IF NOT EXISTS Data (field STRING NOT NULL PRIMARY KEY, value text);")
	return err
}
