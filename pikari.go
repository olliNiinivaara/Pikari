package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const tf = "01-02 15:04"

var mutex sync.Mutex

var appdir, exedir, port, password = "", "", 0, ""

type wsdata struct {
	Sender      string   `json:"sender"`
	Password    string   `json:"password,omitempty"`
	Receivers   []string `json:"receivers,omitempty"`
	Messagetype string   `json:"messagetype"`
	Message     string   `json:"message"`
}

var upgrader = websocket.Upgrader{}

func favicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	fmt.Fprint(w, "data:image/x-icon;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQEAYAAABPYyMiAAAABmJLR0T///////8JWPfcAAAACXBIWXMAAABIAAAASABGyWs+AAAAF0lEQVRIx2NgGAWjYBSMglEwCkbBSAcACBAAAeaR9cIAAAAASUVORK5CYII=\n\n")
}

func pikarijs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	fmt.Fprint(w, pikari)
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

	theuser := user{id: userid, conn: c}
	addUser(&theuser)
	defer removeUser(&theuser, true)

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Pikari server error - web socket read failed:" + err.Error())
			break
		}
		data := wsdata{}
		err = json.Unmarshal(msg, &data)
		if err != nil {
			log.Println("Pikari server error - ws parsing error: " + err.Error())
			break
		}
		if !checkUser(&theuser, data.Password) {
			break
		}
		switch data.Messagetype {
		case "log":
			log.Println(&data.Message)
		case "start":
			//changemessage := changestruct{"start", make(map[string]json.RawMessage)}
			changemessage := changestruct{"start", make(map[string]string)}
			//changemessage.Fields["jokukenttä"] = json.RawMessage(`{"arvo1": "joppis"}`)
			changemessage.Fields["jokukenttä"] = `{"arvo1": "joppis"}`
			jsonresponse, err := json.Marshal(changemessage)
			if err != nil {
				log.Fatal("Pikari server error - data parsing error at start: " + err.Error())
			}
			transmitMessage(&wsdata{"server", "", []string{theuser.id}, "change", string(jsonresponse)}, true)
		case "message":
			transmitMessage(&data, true)
		case "commit":
			commit(&theuser, &data.Message)
		case "rollback":
			rollback(&theuser, true)
		default:
			log.Println("Pikari server error - web socket message type unknown: " + data.Messagetype)
		}
	}
}

func main() {
	flag.IntVar(&port, "port", 8080, "http service port")
	addr := "127.0.0.1:" + strconv.Itoa(port)
	_, callerFile, _, _ := runtime.Caller(0)
	exedir = filepath.Dir(callerFile)
	flag.StringVar(&appdir, "appdir", "", "path to application, absolute or relative to "+exedir)
	flag.StringVar(&password, "password", "", "password for the application")
	flag.Parse()
	if len(appdir) == 0 {
		fmt.Println("Give path to application with appdir parameter, like this: pikari -appdir:myapplication")
		os.Exit(1)
	}
	if !filepath.IsAbs(appdir) {
		appdir = exedir + string(filepath.Separator) + appdir + string(filepath.Separator)
	}
	_, err := os.Stat(appdir)
	if os.IsNotExist(err) {
		fmt.Println("Application directory not found: " + appdir)
		os.Exit(1)
	}
	logfile, err := os.OpenFile(appdir+"pikari.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer logfile.Close()
	log.SetOutput(logfile)
	fmt.Println("Pikari 0.2 starting")
	fs := http.FileServer(http.Dir(appdir))
	http.Handle("/", fs)
	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/ws", ws)
	_, err = os.Stat("pikari.js")
	if os.IsNotExist(err) {
		http.HandleFunc("/pikari.js", pikarijs)
	} else {
		rootfs := http.FileServer(http.Dir(exedir))
		http.Handle("/pikari.js", rootfs)
	}
	http.HandleFunc("/starttransaction", startTransaction)
	fmt.Println("Serving " + appdir + " to " + addr)
	fmt.Println(time.Now().Format(tf) + " users: 0")
	log.Println("---")
	log.Fatal(http.ListenAndServe(addr, nil))
}

const pikari = ``
