package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func openDb(app *appstruct, dir string, maxPagecount int) {
	if maxPagecount <= 0 {
		return
	}
	var err error
	var path = datadir + dir + ".db"
	app.database, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err.Error() + ": " + path)
	}
	app.locks = make(map[string]lock)
	if _, err = app.database.Exec("PRAGMA synchronous = OFF;"); err != nil {
		log.Fatal(err.Error() + ": " + path)
	}
	if _, err = app.database.Exec("PRAGMA journal_mode = OFF;"); err != nil {
		log.Fatal(err)
	}
	if _, err = app.database.Exec("PRAGMA max_page_count = " + strconv.Itoa(maxPagecount) + ";"); err != nil {
		log.Fatal(err)
	}
	if _, err = app.database.Exec("CREATE TABLE IF NOT EXISTS Data (field STRING NOT NULL PRIMARY KEY, value text);"); err != nil {
		log.Fatal(err)
	}
	app.get, err = app.database.Prepare("SELECT field, value FROM data;")
	if err != nil {
		log.Fatal(err)
	}
	app.set, err = app.database.Prepare("INSERT OR REPLACE INTO Data (field, value) VALUES (?,?);")
	if err != nil {
		log.Fatal(err)
	}
	app.del, err = app.database.Prepare("DELETE FROM Data WHERE field = ?;")
	if err != nil {
		log.Fatal(err)
	}

	if dir == "admin" {
		if _, err = app.database.Exec(`INSERT OR IGNORE INTO Data(field, value) VALUES('admin', '{"Name":"Admin", "Maxpagecount": 10000, "Autorestart": 1}');`); err != nil {
			log.Fatal(err)
		}
	}
}

func closeDbs() {
	for _, app := range apps {
		closeDb(app)
	}
	fmt.Println("\nGood bye!")
}

func closeDb(app *appstruct) {
	if app.database == nil {
		return
	}
	app.get.Close()
	app.set.Close()
	if _, err := app.database.Exec("VACUUM;"); err != nil {
		log.Println(err)
	}
	if _, err := app.database.Exec("PRAGMA optimize;"); err != nil {
		log.Println(err)
	}
	app.database.Close()
	app.database = nil
}

func getData(app *appstruct) []byte {
	if app.database == nil {
		return []byte("{}")
	}
	if app.buffer.Len() > 0 {
		return app.buffer.Bytes()
	}
	rows, err := app.get.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	app.buffer.WriteString("{")
	for rows.Next() {
		var field []byte
		var value []byte
		err = rows.Scan(&field, &value)
		if err != nil {
			log.Fatal(err)
		}
		jfield, _ := json.Marshal(string(field))
		app.buffer.Write(jfield)
		app.buffer.WriteString(`:`)
		app.buffer.Write(value)
		app.buffer.WriteString(",")
	}
	if app.buffer.Len() > 1 {
		app.buffer.Truncate(app.buffer.Len() - 1)
	}
	app.buffer.WriteString("}")
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return app.buffer.Bytes()
}

func update(app *appstruct, tx *sql.Tx, field string, value string) bool {
	var err error
	if value == "null" {
		_, err = tx.Stmt(app.del).Exec(field)
	} else {
		_, err = tx.Stmt(app.set).Exec(field, value)
	}
	if err != nil {
		if app.Autorestart == 1 || (app.Autorestart == -1 && apps["admin"].Autorestart == 1) {
			tx.Rollback()
			app.Unlock()
			if dropDb(app) == nil {
				transmitMessage(app, &wsdata{Sender: "server", Receivers: []string{}, Messagetype: "change", Message: "{}"})
				return false
			}
		}
		log.Fatal(err)
	}
	return true
}

func dropDb(app *appstruct) error {
	if app.Name == "Admin" {
		return nil
	}
	_, err := app.database.Exec("DROP TABLE Data")
	_, err = app.database.Exec("CREATE TABLE IF NOT EXISTS Data (field STRING NOT NULL PRIMARY KEY, value text);")
	app.buffer.Reset()
	log.Println("Database dropped")
	return err
}
