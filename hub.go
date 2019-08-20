package main

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type wsdata struct {
	User        string `json:"user"`
	Password    string `json:"password,omitempty"`
	Messagetype string `json:"messagetype"`
	Message     string `json:"message"`
}

type user struct {
	id   string
	conn *websocket.Conn
}

var users = make(map[string]*user)

var mutex sync.Mutex

func addUser(u *user) {
	mutex.Lock()
	if _, ok := users[u.id]; ok {
		removeUser(u, false)
		if dev {
			fmt.Println("Signed-in user kicked out: " + u.id)
		}
	}
	users[u.id] = u
	if dev {
		fmt.Println("User signed in: " + u.id)
		fmt.Println("Users on-line: " + strconv.Itoa(len(users)))
	}
	mutex.Unlock()
}

func removeUser(u *user, lock bool) {
	if lock {
		mutex.Lock()
		defer mutex.Unlock()
	}
	delete(users, u.id)
	u.conn.Close()
	if dev {
		fmt.Println("User signed out: " + u.id)
		fmt.Println("Users on-line: " + strconv.Itoa(len(users)))
	}
}

func respond(receivers *[]string, message *[]byte) {
	mutex.Lock()
	defer mutex.Unlock()
	if len(*receivers) == 0 {
		for _, receiver := range users {
			err := receiver.conn.WriteMessage(websocket.TextMessage, *message)
			if err != nil {
				removeUser(receiver, false)
			}
		}
		return
	}
	for _, receivername := range *receivers {
		if receiver, ok := users[receivername]; ok {
			err := receiver.conn.WriteMessage(websocket.TextMessage, *message)
			if err != nil {
				removeUser(receiver, false)
			}
		}
	}
}