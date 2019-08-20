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

	"github.com/gorilla/websocket"
)

var port, addr, path, exePath, application, dev = 0, "", "", "", "", false

var upgrader = websocket.Upgrader{}

func favicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	fmt.Fprint(w, "data:image/x-icon;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQEAYAAABPYyMiAAAABmJLR0T///////8JWPfcAAAACXBIWXMAAABIAAAASABGyWs+AAAAF0lEQVRIx2NgGAWjYBSMglEwCkbBSAcACBAAAeaR9cIAAAAASUVORK5CYII=\n\n")
}

func handleMessage(message *wsdata) {
	type messagedata struct {
		Receivers []string `json:"receivers"`
		Message   string   `json:"message"`
	}
	m := messagedata{}
	err := json.Unmarshal([]byte(message.Message), &m)
	if err != nil {
		log.Println("Pikari server error - message parsing error: " + err.Error())
		return
	}
	response := &wsdata{User: message.User, Messagetype: message.Messagetype, Message: message.Message}
	jsonresponse, err := json.Marshal(response)
	if err != nil {
		log.Println("Pikari server error - message parsing error : " + err.Error())
	}
	respond(&m.Receivers, &jsonresponse)
}

func ws(w http.ResponseWriter, r *http.Request) {
	userid := r.URL.Query().Get("user")
	if userid == "" {
		log.Fatal("Pikari server error - user name missing in web socket handshake")
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
		switch data.Messagetype {
		case "log":
			log.Println(data.Message)
		case "message":
			handleMessage(&data)
		default:
			log.Println("Pikari server error - web socket read message type unknown: " + data.Messagetype)
		}
	}
}

func main() {
	flag.IntVar(&port, "port", 8080, "http service port")
	addr = "127.0.0.1:" + strconv.Itoa(port)
	flag.StringVar(&application, "app", "HelloWorld", "subdirectory of the application")
	_, callerFile, _, _ := runtime.Caller(0)
	flag.BoolVar(&dev, "dev", false, "set internal development mode")
	flag.Parse()
	exePath = filepath.Dir(callerFile)
	path = exePath + string(filepath.Separator) + application + string(filepath.Separator)
	logfile, err := os.OpenFile(path+"Pikari.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)
	fmt.Println("Pikari 0.1 starting from " + path)
	if dev {
		fmt.Println("Development mode set!")
	}
	initAssets()
	fs := http.FileServer(http.Dir(path))
	http.Handle("/", fs)
	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/ws", ws)
	if dev {
		rootfs := http.FileServer(http.Dir(exePath))
		http.Handle("/pikari.js", rootfs)
	} else {
		http.HandleFunc("/pikari.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			fmt.Fprintf(w, pikari)
		})
	}
	fmt.Println("Serving " + application + " to " + addr)
	log.Println("---")
	log.Fatal(http.ListenAndServe(addr, nil))
}
