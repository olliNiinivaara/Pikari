package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type datastructure struct {
	Description string `json:"description"`
	Fields map[string]string `json:"fields"`
}

type transactionrequest struct {
	User     string   `json:"user"`
	Password string   `json:"password"`
	Fields   []string `json:"fields"`
}

type fieldstruct struct {
	lockedby    *user
	lockedsince string
}

var data = make(map[string]string)
var lockedfields = make(map[string]fieldstruct)

func startTransaction(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request transactionrequest
	err := decoder.Decode(&request)
	if err != nil {
		log.Println("Pikari server error - transaction request parsing error: " + err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/json")
	if len(request.Fields) == 0 {
		w.Write([]byte("'error':'No field(s) defined for transaction'"))
	} else {
		w.Write(*lockFields(request))
	}
}

func getLocks() *[]byte {
	s := make([]string, 0, len(lockedfields))
	for f := range lockedfields {
		s = append(s, f)
	}
	jsonresponse, err := json.Marshal(s)
	if err != nil {
		log.Fatal("Pikari server error - could not jsonify lockedfields")
	}
	return &jsonresponse
}

func lockFields(request transactionrequest) *[]byte {
	if !checkUserstring(request.User, request.Password) {
		s := []byte("No credentials")
		return &s
	}
	mutex.Lock()
	defer mutex.Unlock()
	for i := range lockedfields {
		if lockedfields[i].lockedby.id == request.User {
			return getLocks()
		}
	}
	for _, field := range request.Fields {
		if lockedfield, ok := lockedfields[field]; ok {
			if !wasUserdead(lockedfield.lockedby) {
				return getLocks()
			}
		}
	}
	for _, f := range request.Fields {
		lockedfields[f] = fieldstruct{users[request.User], time.Now().String()}
	}
	var ok = []byte("{}")
	return &ok
}

func commit(u *user, newdata *string) {
	var request datastructure
	err := json.Unmarshal([]byte(*newdata), &request)
	if err != nil {
		log.Println("Pikari server error - could not unmarshal commit data: " + string(*newdata))
		rollback(u, true)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	for field := range request.Fields {
		var lockedfield fieldstruct
		var ok bool
		if lockedfield, ok = lockedfields[field]; !ok {
			delete(request.Fields, field)
		}
		if lockedfield.lockedby != u {
			delete(request.Fields, field)
		}
	}
	defer rollback(u, false)
	if len(request.Fields) == 0 {
		return
	}
	jsonresponse, err := json.Marshal(request)
	if err != nil {
		log.Println("Pikari server error - could not marshal commit data: " + err.Error())
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
	for field := range request.Fields {
		data[field] = request.Fields[field]
	}
	transmitMessage(&wsdata{Sender: u.id, Receivers: []string{}, Messagetype: "change", Message: string(jsonresponse)}, false)
}

func rollback(u *user, lock bool) {
	if lock {
		mutex.Lock()
		defer mutex.Unlock()
	}
	for i := range lockedfields {
		if lockedfields[i].lockedby == u {
			delete(lockedfields, i)
		}
	}
}
