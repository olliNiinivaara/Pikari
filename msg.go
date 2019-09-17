package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

func transmitMessage(message *wsdata, lock bool) {
	var receivers = message.Receivers
	message.Password = ""
	message.Receivers = []string{}
	jsonresponse, err := json.Marshal(message)
	if err != nil {
		log.Println("Pikari server error - message parsing error : " + err.Error())
	}

	if lock {
		mutex.Lock()
		defer mutex.Unlock()
	}
	if len(receivers) == 0 {
		for _, receiver := range users {
			err := receiver.conn.WriteMessage(websocket.TextMessage, jsonresponse)
			if err != nil {
				removeUser(receiver, false)
			}
		}
		return
	}
	for _, receivername := range receivers {
		if receiver, ok := users[" <🏆> "+receivername]; ok {
			err := receiver.conn.WriteMessage(websocket.TextMessage, jsonresponse)
			if err != nil {
				removeUser(receiver, false)
			}
		}
	}
}
