package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
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
	theuser := user{id: userid, conn: c}
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
	mutex.Lock()
	transmitMessage(&wsdata{"server", "", []string{theuser.id}, "start", string(getData())}, false)
	mutex.Unlock()
}

func main() {
	flag.IntVar(&port, "port", 8080, "http service port")
	addr := "127.0.0.1:" + strconv.Itoa(port)
	_, callerFile, _, _ := runtime.Caller(0)
	exedir = filepath.Dir(callerFile) + string(filepath.Separator)
	flag.StringVar(&appdir, "appdir", "", "path to application, absolute or relative to "+exedir)
	flag.StringVar(&password, "password", "", "password for the application")
	flag.Parse()
	if len(appdir) == 0 {
		fmt.Println("Give path to application with appdir parameter, like this: pikari -appdir Nameofmyapplication")
		os.Exit(1)
	}
	if !filepath.IsAbs(appdir) {
		appdir = exedir + appdir + string(filepath.Separator)
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
	fmt.Println(time.Now().Format(tf) + " Pikari 0.5 starting")
	openDb()
	getData()
	fs := http.FileServer(http.Dir(appdir))
	http.Handle("/", fs)
	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/ws", ws)
	http.HandleFunc("/setlocks", setLocks)
	_, err = os.Stat(exedir + "pikari.js")
	if os.IsNotExist(err) {
		createPikariJs()
	}
	rootfs := http.FileServer(http.Dir(exedir))
	http.Handle("/pikari.js", rootfs)
	fmt.Println("Serving " + appdir + " to " + addr)
	fmt.Println("Send SIGINT (Ctrl+C) to quit")
	fmt.Print(time.Now().Format(tf) + " users: 0" + " ")
	log.Println("---")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		closeDb()
		os.Exit(0)
	}()
	err = http.ListenAndServe(addr, nil)
	fmt.Println(err)
	log.Fatal(err)
}

func createPikariJs() {
	file, err := os.Create(exedir + "pikari.js")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	file.WriteString(pikari)
	file.Close()
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
