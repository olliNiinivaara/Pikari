package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

var mutex sync.Mutex

type wsdata struct {
	Sender      string   `json:"sender"`
	Password    string   `json:"w,omitempty"`
	App         string   `json:"app,omitempty"`
	Receivers   []string `json:"receivers,omitempty"`
	Messagetype string   `json:"messagetype"`
	Message     string   `json:"message"`
}

var upgrader = websocket.Upgrader{}

const icon64 = "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"

var icon []byte

func favicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=7776123")
	w.Write(icon)
}

func ws(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	userid := r.URL.Query().Get("user")
	app := r.URL.Query().Get("app")
	if len(app) > 1 {
		app = app[1 : len(app)-1]
	} else {
		app = ""
	}
	if utf8.RuneCountInString(userid) > 200 {
		userid = string([]rune(userid)[0:200])
	}
	if userid == "" {
		log.Println("Pikari server error - user name missing in web socket handshake")
		return
	}
	if !appExists(&app) {
		log.Println("Pikari server error - invalid app name in web socket handshake: " + app)
		return
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Pikari server error - web socket upgrade failed:" + err.Error())
		return
	}
	theuser := createUser(&userid, &app, c)
	defer removeUser(theuser, true)

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Println("Pikari server error - web socket read failed:" + err.Error())
			}
			break
		}
		request := wsdata{}
		if err = json.Unmarshal(msg, &request); err != nil {
			log.Println("Pikari server error - ws parsing error: " + err.Error())
			break
		}
		if !checkUser(theuser, request.Password) {
			break
		}
		switch request.Messagetype {
		case "log":
			if utf8.RuneCountInString(request.Message) > 10000 {
				request.Message = string([]rune(request.Message)[0:1000])
			}
			log.Println(&request.Message)
		case "start":
			start(theuser)
		case "message":
			transmitMessage(theuser.app, &request, true)
		case "commit":
			commit(theuser, &request.Message)
		case "dropdata":
			dropData(theuser.app, theuser.id)
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
	message := startdata{Db: string(getData(theuser.app)), Users: getUsers()}
	b, _ := json.Marshal(message)
	transmitMessage(theuser.app, &wsdata{"server", "in", "", []string{}, "sign", theuser.id}, false)
	transmitMessage(theuser.app, &wsdata{"server", "", "", []string{theuser.id}, "start", string(b)}, false)
	mutex.Unlock()
}

func transmitMessage(app *appstruct, message *wsdata, lock bool) {
	var receivers = message.Receivers
	message.Receivers = []string{}
	jsonresponse, err := json.Marshal(message)
	if err != nil {
		log.Println("Pikari server error - message parsing error : " + err.Error())
		return
	}

	if lock {
		mutex.Lock()
		defer mutex.Unlock()
	}
	if len(receivers) == 0 {
		for _, receiver := range users {
			if receiver.app != app {
				continue
			}
			if err = receiver.conn.WriteMessage(websocket.TextMessage, jsonresponse); err != nil {
				removeUser(receiver, false)
				err = nil
			}
		}
		return
	}
	for _, receivername := range receivers {
		if receiver, ok := users[receivername]; ok {
			if receiver.app != app {
				continue
			}
			if err = receiver.conn.WriteMessage(websocket.TextMessage, jsonresponse); err != nil {
				removeUser(receiver, false)
				err = nil
			}
		}
	}
}
