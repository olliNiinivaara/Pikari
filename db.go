package main

import (
	"encoding/json"
	"log"
	"time"
)

type transactionrequest struct {
	User     string   `json:"user"`
	Password string   `json:"password"`
	Fields   []string `json:"fields"`
}

type fieldlock struct {
	locker      *user
	Lockedby    string `json:"lockedby"`
	Lockedsince string `json:"lockedsince"`
}

var lockedfields = make(map[string]fieldlock)

func startTransaction(u *user, r *string) {
	var request transactionrequest
	err := json.Unmarshal([]byte(*r), &request)
	if err != nil {
		log.Println("Pikari server error - transaction request parsing error: " + err.Error())
		return
	}
	if len(request.Fields) == 0 {
		return
	}
	alreadyLocked := *lockFields(request)
	if len(alreadyLocked) > 2 {
		return
	}
	transmitMessage(&wsdata{Sender: request.User, Receivers: []string{}, Messagetype: "transaction", Message: string(*getLocks(request.Fields))}, true)
}

func getLocks(fields []string) *[]byte {
	var jsonresponse []byte
	var err error
	if fields == nil {
		jsonresponse, err = json.Marshal(lockedfields)
	} else {
		locks := make(map[string]fieldlock)
		for f := range lockedfields {
			if contains(fields, f) {
				locks[f] = lockedfields[f]
			}
		}
		jsonresponse, err = json.Marshal(locks)
	}
	if err != nil {
		log.Fatal("Pikari server error - could not jsonify lockedfields")
	}
	return &jsonresponse
}

func lockFields(request transactionrequest) *[]byte {
	if !checkUserstring(request.User, request.Password) {
		s := []byte("{'error': 'No credentials'}")
		return &s
	}
	mutex.Lock()
	defer mutex.Unlock()
	for i := range lockedfields {
		if lockedfields[i].locker.id == request.User {
			return getLocks(nil)
		}
	}
	for _, field := range request.Fields {
		if lockedfield, ok := lockedfields[field]; ok {
			if !wasUserdead(lockedfield.locker) {
				return getLocks(nil)
			}
		}
	}
	for _, f := range request.Fields {
		lockedfields[f] = fieldlock{users[request.User], request.User, time.Now().UTC().Format(time.RFC3339)}
	}
	var ok = []byte("{}")
	return &ok
}

func removeLocks(u *user) {
	for f := range lockedfields {
		if lockedfields[f].locker == u {
			delete(lockedfields, f)
		}
	}
}

func commit(u *user, newdata *string) {
	type indata struct {
		Description string            `json:"description"`
		Fields      map[string]string `json:"fields"`
	}
	var request indata
	err := json.Unmarshal([]byte(*newdata), &request)
	if err != nil {
		log.Println("Pikari server error - could not unmarshal commit data: " + string(*newdata))
		rollback(u, true)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	for field := range request.Fields {
		var lockedfield fieldlock
		var ok bool
		if lockedfield, ok = lockedfields[field]; !ok {
			delete(request.Fields, field)
		}
		if lockedfield.locker != u {
			delete(request.Fields, field)
		}
	}
	defer removeLocks(u)
	if len(request.Fields) == 0 {
		return
	}
	tx, err := database.Begin()
	if err != nil {
		log.Fatal("Pikari server error - could not start transaction: " + err.Error())
	}
	for field := range request.Fields {
		err = update(tx, field, request.Fields[field])
		if err != nil {
			break
		}
	}
	if err != nil {
		log.Println("Pikari server error - could not commit data: " + err.Error())
		tx.Rollback()
		rollback(u, false)
		return
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal("Pikari server error - could not commit data: " + err.Error())
	}
	buffer.Reset()
	transmitMessage(&wsdata{Sender: u.id, Receivers: []string{}, Messagetype: "change", Message: *newdata}, false)
}

func rollback(u *user, lock bool) {
	if lock {
		mutex.Lock()
		defer mutex.Unlock()
	}
	locks := make(map[string]fieldlock)
	for f := range lockedfields {
		if lockedfields[f].locker == u {
			locks[f] = lockedfields[f]
			delete(lockedfields, f)
		}
	}
	jsonresponse, err := json.Marshal(locks)
	if err != nil {
		log.Fatal("Pikari server error - could not marshal rollbacking locks")
	}
	transmitMessage(&wsdata{Sender: u.id, Receivers: []string{}, Messagetype: "rollback", Message: string(jsonresponse)}, false)
}
