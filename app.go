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
	Maxpagecount int
	Autorestart  int
	Source       string
	Gitbranch    string
	Passive      int
	exists       bool
	database     *sql.DB
	get          *sql.Stmt
	set          *sql.Stmt
	del          *sql.Stmt
	locks        map[string]lock
	buffer       bytes.Buffer
	usercount    int
}

var apps = make(map[string]*appstruct)
var indexbuffer bytes.Buffer

//var admin *appstruct
var appmutex sync.Mutex

func initApps(adminpassword string) {
	var a = new(appstruct)
	openDb(a, "admin", 10000)
	if err := json.Unmarshal(getData(a), &apps); err != nil {
		log.Fatal(err)
	}
	theadmin := apps["admin"]
	theadmin.database = a.database
	theadmin.get = a.get
	theadmin.set = a.set
	theadmin.del = a.del
	theadmin.locks = make(map[string]lock)
	theadmin.Password = adminpassword

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
			updateAdmindata(name, string(b))
		}
	}
	for dir, app := range apps {
		if !app.exists {
			delete(apps, dir)
			updateAdmindata(dir, "null")
		}
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
		maxpagecount := apps["admin"].Maxpagecount
		if apps[*dir].Maxpagecount > -1 {
			maxpagecount = apps[*dir].Maxpagecount
		}
		openDb(app, *dir, maxpagecount)
	}
	app.usercount++
	return app
}

func decrementUsercount(app *appstruct) {
	appmutex.Lock()
	defer appmutex.Unlock()
	app.usercount--
	if app.usercount == 0 && app.Name != "Admin" {
		closeDb(app)
	}
}

func closeApp(dir string) {
	if dir == "admin" {
		log.Println("Bugger, someone tried to close admin")
		return
	}
	removeAllUsers(apps[dir])
	closeDb(apps[dir])
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
