package main

import (
	"bytes"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type user struct {
	id    string
	conn  *websocket.Conn
	since time.Time
	app   *appstruct
}

var users = make(map[string]*user)

func addUser(u *user) {
	mutex.Lock()
	if existinguser, ok := users[u.id]; ok {
		removeUser(existinguser, false)
	}
	users[u.id] = u
	/*if config.Usercount {
		fmt.Print("\r" + time.Now().Format(tf) + " users: " + strconv.Itoa(len(users)) + " ")
	}*/
	mutex.Unlock()
}

func removeUser(u *user, lock bool) {
	if lock {
		mutex.Lock()
		defer mutex.Unlock()
	}
	currentuser, ok := users[u.id]
	if !ok || u != currentuser {
		return
	}
	delete(users, u.id)
	removeLocks(u, true)
	u.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	transmitMessage(&wsdata{"server", "", "", []string{}, "sign", u.id}, false)
	/*if config.Usercount {
		fmt.Print("\r" + time.Now().Format(tf) + " users: " + strconv.Itoa(len(users)) + " ")
	}*/
}

func wasUserdead(u *user) bool {
	if err := u.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
		removeUser(u, false)
		return true
	}
	return false
}

func checkUser(u *user, pw string) bool {
	if _, ok := users[u.id]; !ok {
		log.Println("removed user still on-line: " + u.id)
		return false
	}
	if u.app != nil && pw != u.app.Password {
		log.Println("wrong password: " + u.id)
		return false
	}
	return true
}

func getUser(uid string, pw string) *user {
	var u *user
	var ok bool
	if u, ok = users[uid]; !ok {
		log.Println("unsigned user detected: " + uid)
		return nil
	}
	if pw != u.app.Password {
		log.Println("wrong password: " + uid)
		return nil
	}
	return u
}

func getUsers() string {
	var userstring bytes.Buffer
	userstring.WriteString("{")
	for _, u := range users {
		if !wasUserdead(u) {
			jid, _ := json.Marshal(u.id)
			userstring.WriteString(string(jid) + ":" + strconv.FormatInt(u.since.Unix(), 10) + ",")
		}
	}
	userstring.Truncate(userstring.Len() - 1)
	userstring.WriteString("}")
	return userstring.String()
}
