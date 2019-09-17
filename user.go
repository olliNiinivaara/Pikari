package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type user struct {
	id   string
	conn *websocket.Conn
}

var users = make(map[string]*user)

func addUser(u *user) {
	mutex.Lock()
	if existinguser, ok := users[u.id]; ok {
		removeUser(existinguser, false)
	}
	users[u.id] = u
	fmt.Print("\r" + time.Now().Format(tf) + " users: " + strconv.Itoa(len(users)) + " ")
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
	fmt.Print("\r" + time.Now().Format(tf) + " users: " + strconv.Itoa(len(users)) + " ")
}

func wasUserdead(u *user) bool {
	err := u.conn.WriteMessage(websocket.PingMessage, []byte{})
	if err != nil {
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
	if pw != password {
		log.Println("wrong password: " + u.id)
		return false
	}
	return true
}

func getUser(uid string, pw string) *user {
	if pw != password {
		log.Println("wrong password: " + uid)
		return nil
	}
	var u *user
	var ok bool
	if u, ok = users[uid]; !ok {
		log.Println("unsigned user detected: " + uid)
		return nil
	}
	return u
}

func getUsername(uid string) string {
	name := []rune(uid)
	name = name[5:len(name)]
	return string(name)
}
