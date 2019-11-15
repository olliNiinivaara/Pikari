package main

import (
	"encoding/json"
	"log"
	"net/http"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

type wsdata struct {
	Sender      string   `json:"sender"`
	Password    string   `json:"w,omitempty"`
	App         string   `json:"app,omitempty"`
	Receivers   []string `json:"receivers,omitempty"`
	Messagetype string   `json:"messagetype"`
	Message     string   `json:"message"`
}

var upgrader = websocket.Upgrader{}

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
	if theuser == nil {
		log.Println("App disabled or nonexisting: " + app)
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "No such app"))
		return
	}
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
		if request.Messagetype == "start" {
			start(theuser, request.Password)
			continue
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
		case "message":
			theuser.app.Lock()
			defer theuser.app.Unlock()
			transmitMessage(theuser.app, &request)
		case "commit":
			commit(theuser, &request.Message)
		case "logout":
			c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		default:
			log.Println("Pikari server error - web socket message type unknown: " + request.Messagetype)
		}
	}
}

func start(theuser *user, password string) {
	type startdata struct {
		Db    string
		Users string
	}
	if theuser.app == nil {
		message := startdata{Db: getIndexData(), Users: "{}"}
		b, _ := json.Marshal(message)
		transmitMessage(nil, &wsdata{"server", "", "", []string{theuser.id}, "start", string(b)})
		return
	}
	if password != theuser.app.Password {
		if password == "" {
			transmitMessage(theuser.app, &wsdata{"server", "", "", []string{theuser.id}, "start", "passwordrequired"})
		} else {
			transmitMessage(theuser.app, &wsdata{"server", "", "", []string{theuser.id}, "start", "wrongpassword"})
		}
		return
	}
	theuser.app.Lock()
	defer theuser.app.Unlock()
	message := startdata{Db: string(getData(theuser.app)), Users: getUsers(theuser.app)}
	b, _ := json.Marshal(message)
	transmitMessage(theuser.app, &wsdata{"server", "in", "", []string{}, "sign", theuser.id})
	transmitMessage(theuser.app, &wsdata{"server", "", "", []string{theuser.id}, "start", string(b)})
}

func transmitMessage(app *appstruct, message *wsdata) {
	var receivers = message.Receivers
	message.Receivers = []string{}
	jsonresponse, err := json.Marshal(message)
	if err != nil {
		log.Println("Pikari server error - message parsing error : " + err.Error())
		return
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
