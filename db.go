package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type changestruct struct {
	Description string `json:"description"`
	//Fields      map[string]json.RawMessage `json:"fields"`
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

func commit(u *user, data *string) {
	var request changestruct
	err := json.Unmarshal([]byte(*data), &request)
	if err != nil {
		log.Println("Pikari server error - could not unmarshal commit data: " + string(*data))
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
