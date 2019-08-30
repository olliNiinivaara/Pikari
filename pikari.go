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

/*func pikarijs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	fmt.Fprint(w, pikari)
}*/

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
		case "transaction":
			startTransaction(&theuser, &request.Message)
		case "commit":
			commit(&theuser, &request.Message)
		case "rollback":
			rollback(&theuser, true)
		default:
			log.Println("Pikari server error - web socket message type unknown: " + request.Messagetype)
		}
	}
}

func start(theuser *user) {
	type startdata struct {
		Description string `json:"description"`
		Fields      string `json:"fields"`
	}
	mutex.Lock()
	response, err := json.Marshal(startdata{"start", string(getData())})
	if err != nil {
		log.Fatal("Pikari server error - data parsing error: " + err.Error())
	}
	transmitMessage(&wsdata{"server", "", []string{theuser.id}, "change", string(response)}, false)
	mutex.Unlock()
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
	fmt.Println("Pikari 0.3 starting")
	openDb()
	defer closeDb()
	getData()
	fs := http.FileServer(http.Dir(appdir))
	http.Handle("/", fs)
	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/ws", ws)
	_, err = os.Stat("pikari.js")
	if os.IsNotExist(err) {
		createPikariJs()
	}
	rootfs := http.FileServer(http.Dir(exedir))
	http.Handle("/pikari.js", rootfs)
	fmt.Println("Serving " + appdir + " to " + addr)
	fmt.Print(time.Now().Format(tf) + " users: 0" + " ")
	log.Println("---")
	err = http.ListenAndServe(addr, nil)
	fmt.Println(err)
	log.Fatal(err)
}

func createPikariJs() {
	file, err := os.Create("pikari.js")
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

const pikari = `
/**
 * @file Pikari API
 * @see https://github.com/olliNiinivaara/Pikari
 * @author Olli Niinivaara
 * @copyright Olli Niinivaara 2019
 * @license MIT
 * @version 0.3
 */


/** @namespace
 * @description the global Pikari object. To initialize, add listeners and call start.
 * @global
 * @example Pikari.addCommitListener(function(description, sender, fields) { do something }); Pikari.start("John")
*/
window.Pikari = new Object()


/** 
 *  @description Local copy of the whole database. Changes committed to database by any user are automatically updated to it.
 *  @example Pikari.data[Pikari.userdata] = "some data"
 */
Pikari.data = {}


/**
 @typedef Pikari.Lockedfield
 @type {Object}
 @property {string} lockedby {@link Pikari.userdata} identifier of user who started the transaction 
 @property {string} lockedsince The start time of the transaction
 */

/** 
 * @description Database fields that are currently locked due to ongoing transactions. Contains properties of type  {@link Pikari.Lockedfield}
 * @example console.log(Object.keys(Pikari.fieldsInTransaction))
 */
Pikari.fieldsInTransaction = {}


/** 
 * @description Name of the user.
 */
Pikari.user = "Anon-"+ Math.floor(Math.random() * Number.MAX_SAFE_INTEGER)


/** 
 * @description Name of the user's personal database field (simply user with a leading @ - to prevent name clashes)
 */
Pikari.userdata = "@" + Pikari.user


/**
 * Helper function to strip the leading @
 * @param {string} userdata - A user identifier (starting with @)
 * @returns {string} User's name
 */
Pikari.getUsername = function(userdata) {
  return userdata.substring(1)
}


Pikari.start = function(user, password) {
  if (user) {
    Pikari.user = String(user)
    Pikari.userdata = "@" + Pikari.user
  }
  if (!password) password = ""
  Pikari._password = String(password)
  Pikari._startWebSocket()
}

Pikari.addTransactionListener = function(handler) {
  Pikari._transactionlisteners.push(handler)
}

Pikari.addCommitListener = function(handler) {
  Pikari._commitlisteners.push(handler)
}

Pikari.addRollbackListener = function(handler) {
  Pikari._rollbacklisteners.push(handler)
}

Pikari.addMessageListener = function(handler) {
  Pikari._messagelisteners.push(handler)
}

Pikari.log = function(event) {
  Pikari._sendToServer("log", event)
}

Pikari.sendMessage = function(message, receivers) {
  if (!receivers) receivers = []
  Pikari._sendToServer("message", JSON.stringify({"receivers": receivers, "message": message}))
}

Pikari.startTransaction = function(fields) {
  if (Pikari._fieldsInThisTransaction) throw("Transaction is already active")
  if (!fields || (Array.isArray(fields) && fields.length == 0)) fields = Pikari._getAllFields()
  if (!Array.isArray(fields)) fields = [String(fields)]
  const request = {"user":Pikari.userdata, "fields":fields}
  Pikari._oldData = JSON.parse(JSON.stringify(Pikari.data))
  Pikari._sendToServer("transaction", JSON.stringify(request))
}

Pikari.commit = function(description) {
  if (!Pikari._fieldsInThisTransaction) throw("No transaction to commit")
  if (description == null) description = ""
  let fields = {}  
  Object.keys(Pikari._fieldsInThisTransaction).forEach(f => {fields[f] = JSON.stringify(Pikari.data[f])})
  Pikari._sendToServer("commit", JSON.stringify({"description": description, "fields": fields}))
}

Pikari.rollback = function() {
  if (!Pikari._fieldsInThisTransaction) throw("No transaction to rollback")
  Pikari._sendToServer("rollback")
}


//private stuff-------------------------------------------

Pikari._transactionlisteners = []
Pikari._commitlisteners = []
Pikari._rollbacklisteners = []
Pikari._messagelisteners = []
Pikari._fieldsInThisTransaction = false

Pikari._reportError = function(error) {
  error = "Pikari client error - " + error
  console.log(error)
  if (!error.includes("Web socket problem: ")) Pikari.log(error)
  alert(error)
  throw error
}

Pikari._getAllFields = function() {
  let fields = Object.getOwnPropertyNames(Pikari.data)
  if (fields.length == 0) {
    Pikari.data["db"] = {}
    fields = ["db"]
  }
  return fields
}

Pikari._sendToServer = function(messagetype, message) {
  if (!Pikari._ws) alert ("No connection server!")
  else Pikari._ws.send(JSON.stringify({"sender": Pikari.userdata, "password":Pikari._password, "messagetype": messagetype, "message": message}))
}

Pikari._handleTransaction = function(d) {
  const fields = JSON.parse(d.message)
  Object.assign(Pikari.fieldsInTransaction, fields)
  if (d.sender == Pikari.userdata) {
    Pikari._fieldsInThisTransaction = {}
    Object.assign(Pikari._fieldsInThisTransaction, fields)
  }
  for(let l of Pikari._transactionlisteners) l(d.sender, fields)
}

Pikari._handleCommit = function(d) {
  const dd = JSON.parse(d.message)
  if (dd.description == "start") {
    const ddd = JSON.parse(dd.fields)       
    Object.entries(ddd).forEach(([field, data]) => { Pikari.data[field] = data })
    for(let l of Pikari._commitlisteners) l("start", "", Object.keys(ddd))
  } else {
    Object.entries(dd.fields).forEach(([field, data]) => { Pikari.data[field] = JSON.parse(data) })
    if (d.sender == Pikari.userdata) Pikari._fieldsInThisTransaction = false
    for(let l of Pikari._commitlisteners) l(dd.description, d.sender, Object.keys(dd))
  }
}

Pikari._handleRollback = function(d) {
  const fields = JSON.parse(d.message)
  Object.keys(fields).forEach(f => { delete Pikari.fieldsInTransaction[f] })
  if (d.sender == Pikari.userdata) {
    Pikari._fieldsInThisTransaction = false
    Pikari.data = JSON.parse(JSON.stringify(Pikari._oldData))
  }
  for(let l of Pikari._rollbacklisteners) l(d.sender, fields)
}

Pikari._startWebSocket = function() {
  Pikari._ws = new WebSocket("ws://"+document.location.host+"/ws?user="+Pikari.userdata)

  Pikari._ws.onopen = function() {  
    Pikari._sendToServer("start")
  }

  Pikari._ws.onclose = function() {
    alert("Connection to Pikari server was lost!")
    Pikari._ws = null
    Pikari.data = {}
    for(let l of Pikari._commitlisteners) l("stop", "", null)
  }

  Pikari._ws.onmessage = function(evt) {
    const d = JSON.parse(evt.data)
    switch (d.messagetype) {
      case "message": { for(let l of Pikari._messagelisteners) l(d.sender, d.message); break }
      case "transaction":  { Pikari._handleTransaction(d); break }      
      case "change": { Pikari._handleCommit(d); break }
      case "rollback":  { Pikari._handleRollback(d); break }
      default: Pikari._reportError("Unrecognized message type received: " + d.messagetype)
    }
  }

  Pikari._ws.onerror = function(evt) { Pikari._reportError("Web socket problem: " + evt.data) }
}
`
