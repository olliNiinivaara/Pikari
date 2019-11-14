package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type lockrequest struct {
	User     string   `json:"user"`
	Password string   `json:"w"`
	Locks    []string `json:"locks"`
}

type lock struct {
	locker      *user
	Lockedby    string `json:"lockedby"`
	Lockedsince string `json:"lockedsince"`
}

func setLocks(w http.ResponseWriter, r *http.Request) {
	var request lockrequest
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(b, &request); err != nil {
		log.Println("Pikari server error - setLocks parsing error: " + err.Error())
		w.Write([]byte(`{"error": "invalid setLocks request"}`))
	} else {
		var theuser = getUser(request.User, request.Password)
		if theuser == nil {
			w.Write([]byte(`{"error": "No credentials"}`))
		} else {
			theuser.app.Lock()
			defer theuser.app.Unlock()
			removeLocks(theuser, false)
			tryToAcquireLocks(theuser, request)
			b, _ := json.Marshal(theuser.app.locks)
			w.Write(b)
			notifyLocking(theuser.app, &theuser.id)
		}
	}
}

func tryToAcquireLocks(u *user, r lockrequest) {
	for _, l := range r.Locks {
		if locked, ok := u.app.locks[l]; ok {
			if locked.locker != u && !wasUserdead(locked.locker) {
				return
			}
		}
	}
	for _, l := range r.Locks {
		u.app.locks[l] = lock{u, u.id, time.Now().UTC().Format(time.RFC3339)}
	}
}

func removeLocks(u *user, notify bool) {
	var trueremoval = false
	for l := range u.app.locks {
		if u.app.locks[l].locker == u {
			delete(u.app.locks, l)
			trueremoval = true
		}
	}
	if notify && trueremoval {
		notifyLocking(u.app, &u.id)
	}
}

func notifyLocking(app *appstruct, sender *string) {
	b, _ := json.Marshal(app.locks)
	transmitMessage(app, &wsdata{Sender: *sender, Receivers: []string{}, Messagetype: "lock", Message: string(b)})
}

func commit(u *user, newdata *string) {
	var fields map[string]string
	if err := json.Unmarshal([]byte(*newdata), &fields); err != nil {
		log.Println("Pikari server error - could not unmarshal commit data: " + string(*newdata))
		return
	}
	u.app.Lock()
	defer u.app.Unlock()
	defer removeLocks(u, true)
	if len(fields) == 0 {
		return
	}
	tx, err := u.app.database.Begin()
	if err != nil {
		log.Fatal("Pikari server error - could not start transaction: " + err.Error())
	}
	for field := range fields {
		if ok := update(u.app, tx, field, fields[field]); !ok {
			return
		}
		if u.app.Name == "Admin" {
			updateApp(field, fields[field])
		}
	}
	if err != nil {
		log.Println("Pikari server error - could not commit data: " + err.Error())
		tx.Rollback()
		return
	}
	if err = tx.Commit(); err != nil {
		log.Fatal("Pikari server error - could not commit data: " + err.Error())
	}
	u.app.buffer.Reset()
	transmitMessage(u.app, &wsdata{Sender: u.id, Receivers: []string{}, Messagetype: "change", Message: *newdata})
}

func dropData(app *appstruct, username string) {
	app.Lock()
	defer app.Unlock()
	app.locks = make(map[string]lock)
	tx, err := app.database.Begin()
	if err != nil {
		log.Fatal("Pikari server error - could not start drop transaction: " + err.Error())
	}
	if err = dropDb(app, tx); err != nil {
		log.Println("Pikari server error - could not drop: " + err.Error())
		tx.Rollback()
		return
	}
	if err = tx.Commit(); err != nil {
		log.Fatal("Pikari server error - could not commit drop: " + err.Error())
	}
	app.buffer.Reset()
	app.buffer.WriteString("{}")
	transmitMessage(app, &wsdata{Sender: username, Receivers: []string{}, Messagetype: "lock", Message: "{}"})
	transmitMessage(app, &wsdata{Sender: username, Receivers: []string{}, Messagetype: "drop", Message: "{}"})
}
