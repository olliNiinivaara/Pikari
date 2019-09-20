package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var mutex sync.Mutex

type wsdata struct {
	Sender      string   `json:"sender"`
	Password    string   `json:"w,omitempty"`
	Receivers   []string `json:"receivers,omitempty"`
	Messagetype string   `json:"messagetype"`
	Message     string   `json:"message"`
}

var upgrader = websocket.Upgrader{}

func favicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	io.WriteString(w, "data:image/x-icon;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQEAYAAABPYyMiAAAABmJLR0T///////8JWPfcAAAACXBIWXMAAABIAAAASABGyWs+AAAAF0lEQVRIx2NgGAWjYBSMglEwCkbBSAcACBAAAeaR9cIAAAAASUVORK5CYII=\n\n")
}

func ws(w http.ResponseWriter, r *http.Request) {
	userid := r.URL.Query().Get("user")
	if userid == "" {
		log.Println("Pikari server error - user name missing in web socket handshake")
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Pikari server error - web socket upgrade failed:" + err.Error())
		return
	}
	theuser := user{id: userid, conn: c, since: time.Now()}
	addUser(&theuser)
	defer removeUser(&theuser, true)

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Println("Pikari server error - web socket read failed:" + err.Error())
			}
			break
		}
		request := wsdata{}
		err = json.Unmarshal(msg, &request)
		if err != nil {
			log.Println("Pikari server error - ws parsing error: " + err.Error())
			break
		}
		if !checkUser(&theuser, request.Password) {
			break
		}
		switch request.Messagetype {
		case "log":
			log.Println(&request.Message)
		case "start":
			start(&theuser)
		case "message":
			transmitMessage(&request, true)
		case "commit":
			commit(&theuser, &request.Message)
		case "dropdata":
			dropData()
		default:
			log.Println("Pikari server error - web socket message type unknown: " + request.Messagetype)
		}
	}
}

func start(theuser *user) {
	type startdata struct {
		Db    string
		Users string
	}
	mutex.Lock()
	message := startdata{Db: string(getData()), Users: getUsers()}
	b, _ := json.Marshal(message)
	transmitMessage(&wsdata{"server", "in", []string{}, "sign", theuser.id}, false)
	transmitMessage(&wsdata{"server", "", []string{theuser.id}, "start", string(b)}, false)
	mutex.Unlock()
}

func transmitMessage(message *wsdata, lock bool) {
	var receivers = message.Receivers
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
		if receiver, ok := users[receivername]; ok {
			err := receiver.conn.WriteMessage(websocket.TextMessage, jsonresponse)
			if err != nil {
				removeUser(receiver, false)
			}
		}
	}
}
