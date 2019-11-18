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
	globalmutex.Lock()
	defer globalmutex.Unlock()
	if existinguser, ok := users[u.id]; ok {
		removeUser(existinguser, false)
	}
	users[u.id] = u
	/*if config.Usercount {
		fmt.Print("\r" + time.Now().Format(tf) + " users: " + strconv.Itoa(len(users)) + " ")
	}*/
}

func removeAllUsers(app *appstruct) {
	for _, u := range users {
		if u.app == app {
			u.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			delete(users, u.id)
			app.usercount = 0
		}
	}
}

func removeUser(u *user, lock bool) {
	if lock {
		globalmutex.Lock()
		defer globalmutex.Unlock()
	}
	currentuser, ok := users[u.id]
	if !ok || u != currentuser {
		return
	}
	delete(users, u.id)
	u.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if u.app != nil {
		removeLocks(u, true)
		transmitMessage(u.app, &wsdata{"server", "", "", []string{}, "sign", u.id})
		decrementUsercount(u.app)
		u.app = nil
	}
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

func getUsers(app *appstruct) string {
	var userstring bytes.Buffer
	userstring.WriteString("{")
	for _, u := range users {
		if !wasUserdead(u) && app == u.app {
			jid, _ := json.Marshal(u.id)
			userstring.WriteString(string(jid) + ":" + strconv.FormatInt(u.since.Unix(), 10) + ",")
		}
	}
	userstring.Truncate(userstring.Len() - 1)
	userstring.WriteString("}")
	return userstring.String()
}
