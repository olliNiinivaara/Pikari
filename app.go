package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type appstruct struct {
	Name         string
	Password     string
	Usercount    int
	Maxpagecount int
	Autorestart  int
	Gitbranch    string
	Repourl      string
	exists       bool
	database     *sql.DB
	get          *sql.Stmt
	set          *sql.Stmt
	del          *sql.Stmt
	buffer       bytes.Buffer
}

var apps = make(map[string]*appstruct)
var indexbuffer bytes.Buffer
var admin *appstruct
var appmutex sync.Mutex

func initApps(adminpassword string) {
	admin = new(appstruct)
	openDb(admin, "admin", 10000)
	if err := json.Unmarshal(getData(admin), &apps); err != nil {
		log.Fatal(err)
	}
	admin.Name = "Admin"
	admin.Maxpagecount = apps["admin"].Maxpagecount
	admin.Autorestart = apps["admin"].Autorestart
	admin.Password = adminpassword
	apps["admin"] = admin
	os.Mkdir(exedir+"admin", 0700)

	files, err := ioutil.ReadDir(exedir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		name := f.Name()
		a := apps[name]
		if a != nil {
			a.exists = true
		} else {
			apps[name] = &appstruct{Name: name, Maxpagecount: -1, Autorestart: -1, exists: true}
			b, _ := json.Marshal(apps[name])
			updateAdmin(name, string(b))
		}
	}
	for dir, app := range apps {
		if !app.exists {
			delete(apps, dir)
			updateAdmin(dir, "null")
		}
	}
}

func updateAdmin(dir, value string) {
	tx, err := admin.database.Begin()
	if err != nil {
		log.Fatal("Pikari server error - could not start transaction: " + err.Error())
	}
	update(admin, tx, dir, value)
	if err = tx.Commit(); err != nil {
		log.Fatal("Pikari server error - could not commit data: " + err.Error())
	}
}

func createUser(uid *string, dir *string, c *websocket.Conn) *user {
	theuser := user{id: *uid, conn: c, since: time.Now(), app: getApp(dir)}
	addUser(&theuser)
	return &theuser
}

func appExists(dir *string) bool {
	if *dir == "" {
		return true
	}
	return apps[*dir] != nil
}

func getApp(dir *string) *appstruct {
	appmutex.Lock()
	defer appmutex.Unlock()
	if *dir == "" {
		return nil
	}
	app := apps[*dir]
	if app.database == nil {
		maxpagecount := admin.Maxpagecount
		if apps[*dir].Maxpagecount > -1 {
			maxpagecount = apps[*dir].Maxpagecount
		}
		openDb(app, *dir, maxpagecount)
	}
	app.Usercount++
	return app
}

func getIndexData() []byte {
	indexbuffer.Reset()
	indexbuffer.WriteString("{")
	for f, v := range apps {
		jfield, _ := json.Marshal(f)
		indexbuffer.Write(jfield)
		indexbuffer.WriteString(`:`)
		jvalue, _ := json.Marshal(v.Name)
		indexbuffer.Write(jvalue)
		indexbuffer.WriteString(",")
	}
	if indexbuffer.Len() > 1 {
		indexbuffer.Truncate(indexbuffer.Len() - 1)
	}
	indexbuffer.WriteString("}")
	return indexbuffer.Bytes()
}
