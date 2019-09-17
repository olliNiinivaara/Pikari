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
	Password string   `json:"password"`
	Locks    []string `json:"locks"`
}

type lock struct {
	locker      *user
	Lockedby    string `json:"lockedby"`
	Lockedsince string `json:"lockedsince"`
}

var locks = make(map[string]lock)

func setLocks(w http.ResponseWriter, r *http.Request) {
	var request lockrequest
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &request)
	if err != nil {
		log.Println("Pikari server error - setLocks parsing error: " + err.Error())
		w.Write([]byte(`{"error": "invalid setLocks request"}`))
	} else {
		var theuser = getUser(" <🏆> "+request.User, request.Password)
		if theuser == nil {
			w.Write([]byte(`{"error": "No credentials"}`))
		} else {
			mutex.Lock()
			removeLocks(theuser, false)
			tryToAcquireLocks(theuser, request)
			b, _ := json.Marshal(locks)
			w.Write(b)
			notifyLocking(&theuser.id)
			mutex.Unlock()
		}
	}
}

func tryToAcquireLocks(u *user, r lockrequest) {
	for _, l := range r.Locks {
		if locked, ok := locks[l]; ok {
			if locked.locker != u && !wasUserdead(locked.locker) {
				return
			}
		}
	}
	for _, l := range r.Locks {
		locks[l] = lock{u, getUsername(u.id), time.Now().UTC().Format(time.RFC3339)}
	}
}

func removeLocks(u *user, notify bool) {
	var trueremoval = false
	for l := range locks {
		if locks[l].locker == u {
			delete(locks, l)
			trueremoval = true
		}
	}
	if notify && trueremoval {
		notifyLocking(&u.id)
	}
}

func notifyLocking(sender *string) {
	b, _ := json.Marshal(locks)
	transmitMessage(&wsdata{Sender: getUsername(*sender), Receivers: []string{}, Messagetype: "lock", Message: string(b)}, false)
}

func commit(u *user, newdata *string) {
	var fields map[string]string
	err := json.Unmarshal([]byte(*newdata), &fields)
	mutex.Lock()
	defer mutex.Unlock()
	defer removeLocks(u, true)
	if err != nil {
		log.Println("Pikari server error - could not unmarshal commit data: " + string(*newdata))
		return
	}
	if len(fields) == 0 {
		return
	}
	tx, err := database.Begin()
	if err != nil {
		log.Fatal("Pikari server error - could not start transaction: " + err.Error())
	}
	for field := range fields {
		if ok := update(tx, field, fields[field]); !ok {
			return
		}
	}
	if err != nil {
		log.Println("Pikari server error - could not commit data: " + err.Error())
		tx.Rollback()
		return
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal("Pikari server error - could not commit data: " + err.Error())
	}
	buffer.Reset()
	transmitMessage(&wsdata{Sender: getUsername(u.id), Receivers: []string{}, Messagetype: "change", Message: *newdata}, false)
}

func dropData() {
	mutex.Lock()
	defer mutex.Unlock()
	locks = make(map[string]lock)
	tx, err := database.Begin()
	if err != nil {
		log.Fatal("Pikari server error - could not start drop transaction: " + err.Error())
	}
	err = dropDb(tx)
	if err != nil {
		log.Println("Pikari server error - could not drop: " + err.Error())
		tx.Rollback()
		return
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal("Pikari server error - could not commit drop: " + err.Error())
	}
	buffer.Reset()
	buffer.WriteString("{}")
	transmitMessage(&wsdata{Sender: "server", Receivers: []string{}, Messagetype: "lock", Message: "{}"}, false)
	transmitMessage(&wsdata{Sender: "server", Receivers: []string{}, Messagetype: "change", Message: "{}"}, false)
}
