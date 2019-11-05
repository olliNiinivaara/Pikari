package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func openDb(app *appstruct, dir string, maxPagecount int) {
	if maxPagecount <= 0 {
		return
	}
	var err error
	var path = exedir + dir + string(filepath.Separator) + "data.db"
	app.database, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err.Error() + ": " + path)
	}
	if _, err = app.database.Exec("PRAGMA synchronous = OFF;"); err != nil {
		log.Fatal(err)
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
		if app.database == nil {
			continue
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
	fmt.Println("\nGood bye!")
}

func getData(app *appstruct) []byte {
	if app == nil {
		return getIndexData()
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
		if config.Autorestart {
			tx.Rollback()
			mutex.Unlock()
			dropData(app, "server autorestart")
			return false
		}
		log.Fatal(err)
	}
	return true
}

func dropDb(app *appstruct, tx *sql.Tx) error {
	_, err := app.database.Exec("DROP TABLE Data")
	_, err = app.database.Exec("CREATE TABLE IF NOT EXISTS Data (field STRING NOT NULL PRIMARY KEY, value text);")
	log.Println("Database dropped")
	return err
}
