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
	if _, ok := users[u.id]; ok {
		removeUser(u, false)
	}
	users[u.id] = u
	fmt.Println("\r" + time.Now().Format(tf) + " users: " + strconv.Itoa(len(users)))
	mutex.Unlock()
}

func removeUser(u *user, lock bool) {
	if lock {
		mutex.Lock()
		defer mutex.Unlock()
	}
	rollback(u, false)
	delete(users, u.id)
	fmt.Println("\r" + time.Now().Format(tf) + " users: " + strconv.Itoa(len(users)))
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
		log.Println("wrong password detected")
		removeUser(u, true)
		return false
	}
	return true
}

func checkUserstring(uid string, pw string) bool {
	var theuser *user
	var ok bool
	if theuser, ok = users[uid]; !ok {
		log.Println("unsigned user detected: " + uid)
		return false
	}
	if pw != password {
		log.Println("wrong password detected")
		removeUser(theuser, true)
		return false
	}
	return true
}
