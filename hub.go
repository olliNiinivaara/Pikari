package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	mu   sync.Mutex
}

var users = make(map[string]*user)
var sendmessage = make(chan wsdata)
var register = make(chan *user)

func loopForever() {
	for {
		select {
		case user := <-register:
			users[user.id] = user
		/*case client := <-h.unregister:
		if _, ok := h.clients[client.ID]; ok {
			delete(h.clients, client.ID)
			close(client.send)
		}*/
		case message := <-sendmessage:
			if user, ok := users[message.User]; ok {
				sendMessage(user.conn, user.id, message.Message)
				/*select {
				case user.conn.send <- message.data:
				default:
					close(client.send)
					delete(h.connections, client)
				}*/
			}
		}
	}
}

func sendMessage(c *websocket.Conn, user string, msg string) {
  parseeraukset eri rutiiniin, vain tallennus ja lähetys tänne


	type messagedata struct {
		Receivers []string `json:"receivers"`
		Message   string   `json:"message"`
	}
	m := messagedata{}
	err := json.Unmarshal([]byte(msg), &m)
	if err != nil {
		log.Println("Pikari server error - message parsing error: ", err)
		return
	}
	data := &wsdata{User: user, Messagetype: "message", Message: m.Message}
	jsondata, err := json.Marshal(data)
	if err != nil {
		log.Println("Pikari server error - message parsing error : ", err)
	}
	if len(m.Receivers) == 0 {
		for _, receiver := range users {
			err = receiver.conn.WriteMessage(websocket.TextMessage, []byte(jsondata))
			if err != nil {
				log.Println("Pikari server error - web socket write message failed: ", err)
			}
		}
	}
	for _, receivername := range m.Receivers {
		if receiver, ok := users[receivername]; ok {
			err = receiver.conn.WriteMessage(websocket.TextMessage, []byte(jsondata))
			if err != nil {
				log.Println("Pikari server error - web socket write message failed: ", err)
			}
			/*select {
			case user.conn.send <- message.data:
			default:
				close(client.send)
				delete(h.connections, client)
			}*/
		}
	}
}

/*func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}*/
