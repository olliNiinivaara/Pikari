package main

const pikari = `/**
* @file Pikari API
* @see https://github.com/olliNiinivaara/Pikari
* @author Olli Niinivaara
* @copyright Olli Niinivaara 2019
* @license MIT
* @version 0.9
*/

/** @namespace
* @description The global Pikari object. To initialize, add listeners and call {@link Pikari.start}.
* @global
* 
*/
window.Pikari = new Object()

/** 
* @description Local copy of the database.
* <br>Keys identify fields and values are field values.
* <br>Keys should be strings but values can be any objects, even nested.
* <br>If you want to delete a field from database, set it's value to null.
* <br>Changes to fields committed by any user are automatically updated.
*  @type {Map<string, *>}
*  @example
*  if (await Pikari.setLocks("somefield", someotherfield)) {
*    Pikari.data.set("somefield", "some value")
*    Pikari.data.get(someotherfield)[someproperty] = "some value" 
*    Pikari.commit()
*  }
*/
Pikari.data = new Map()

/** 
* @description Helper function to generate a number that is unique enough for prototyping purposes.
* @return {number} a random integer that is hopefully unique.
*/
Pikari.generateKey = function () {
  return Math.floor(Math.random() * Number.MAX_SAFE_INTEGER)
}

/** 
* @description Name of the user. Automatically generated if not explicitly given as user query param in url or at {@link Pikari.start}.
* Maximum length of a name is 200 letters.
* @type {string}
*/
Pikari.user = "Anon-" + Pikari.generateKey()

/** 
* @description Names of locks that are currently held by the current user.
* <br>Locks can be acquired with {@link Pikari.setLocks}.
* <br>At least one lock must be held before commit can be called.
* @type {string[]}
*/
Pikari.mylocks = []

/**
@typedef Pikari.Lock
@type {Object}
@property {string} lockedby current lock owner 
@property {string} lockedsince The start time of locking
*/
/** 
* @description Locks that are currently held.
* <br>Object's property names (keys) identify the locks and properties are of type {@link Pikari.Lock}.
* <br>Changes to locks can be listened with {@link Pikari.addLockListener}.
* @type {Object.<string, Pikari.Lock>}
*/
Pikari.locks = {}

/** 
* @description Users currenty online.
* <br>Keys identify users and values tell the times when users became online.
* <br>Changes to user presence can be listened with {@link Pikari.addUserListener}.
* <br>The users map is ordered by ascending connection time (which means that {@link Pikari.user} is last at {@link Pikari.EVENT.START})
* @type {Map<string,Date>}
*/
Pikari.users = new Map()

/**
* @description Describes the reason for data change in {@link {changeCallback}}
* @typedef {Object} EVENT
* @memberOf Pikari
* @property {string} START - Connection with server is established (see {@link Pikari.start}) or server restarted itself
* @property {string} COMMIT - Someone committed data (see {@link Pikari.commit})
* @property {string} ROLLBACK - The local user rollbacked (see {@link Pikari.rollback})
*/
Pikari.EVENT = {
  START: "START",
  COMMIT: "COMMIT",
  ROLLBACK: "ROLLBACK"
}
Object.freeze(Pikari.EVENT)

/** 
* @description Helper function to clean a string so that it can be safely used as innerHTML.
* @param {string} str - the string to be cleaned
* @return {string} the cleaned version.
*/
Pikari.clean = function (str) {
  return String(str).trim().replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

/** 
* @description Connects the Pikari client to the Pikari server with given name and password (both optional).
* If a user with same name is already connected, existing user will be immediately disconnected.
* @param {string} user - user name (max 200 letters). If user is false, user name from URL query param is used.
* If URL query param does not exist and user is "prompt", user name is asked. Finally, if no user name is given, a random name is used.
* @param {string} password - if password is set for application, this must match it
*/
Pikari.start = function (user, password) {
  if (!Pikari._startstate) {
    Pikari._startstate = "starting"
    let style = document.createElement('style')
    document.head.appendChild(style)
    style.sheet.insertRule(" body.waiting * { cursor: wait; }", 0)
  }
  Pikari.waiting(true)
  if (!user  || user == "prompt") {
    const userparam = new URLSearchParams(window.location.search).get('user')
    if (userparam) user = userparam
    else if (user == "prompt") user = prompt(document.title+"\nUser name")
  }
  if (user) Pikari.user = user
  Pikari.user = Pikari.user.substring(0, 200)
  if (!password) password = ""
  Pikari._password = btoa(String(password))
  Pikari._startWebSocket()
}

/**
* @callback stopCallback
* @memberOf Pikari
* @description Triggered at most once, only if connection to server is lost (usually might happen when server is closed).
* @return {boolean} - Return true to prevent default handling (which disables all elements and shows an alert).
*/
/**
* @description Set handler for the stop event.
* @param {stopCallback} handler - {@link Pikari.stopCallback}
*/
Pikari.setStopListener = function (handler) {
  Pikari._stoplistener = handler
}

/**
* @callback changeCallback
* @memberOf Pikari
* @description Triggered whenever data might have been changed.
* @param {EVENT} changetype - {@link Pikari.EVENT} 
* @param {string[]} changedfields - an array of changed field names
* @param {string} changer - changer's name ("" if EVENT == START)
*/
/**
* @description Add handler for change events.
* @param {changeCallback} handler - {@link Pikari.changeCallback}
*/
Pikari.addChangeListener = function (handler) {
  Pikari._changelisteners.push(handler)
}

/**
* @callback messageCallback
* @memberOf Pikari
* @description Triggered when a user sends a message.
* @param {string} sender - sender of the message
* @param {string} message - the message itself
*/
/**
* @description Add handler for messages sent by other users.
* @param {messageCallback} handler - {@link Pikari.messageCallback}
*/
Pikari.addMessageListener = function (handler) {
  Pikari._messagelisteners.push(handler)
}

/**
* @callback lockCallback
* @memberOf Pikari
* @description Triggered when locks are acquired or released.
* <br>Note: also local {@link Pikari.setLocks} calls will trigger the lockCallback.
*/
/**
* @description Add handler for changes in lock ownerships.
* @param {lockCallback} handler - {@link Pikari.lockCallback}
*/
Pikari.addLockListener = function (handler) {
  Pikari._locklisteners.push(handler)
}

/**
* @callback userCallback
* @memberOf Pikari
* @description  Triggered when a user logs in or disconnects.
* @param {string} user - the user that logged in or disconnected
* @param {boolean} login - true iff the user logged in
*/
/**
* @description Add handler for changes in on-line users.
* @param {userCallback} handler - {@link Pikari.userCallback}
*/
Pikari.addUserListener = function (handler) {
  Pikari._userlisteners.push(handler)
}

/**
* @description A helper function to get field value or default if undefined.
* The default will be set to local data but not sent to server database.
* @param {string} field - Field name.
* @param {*} defaultvalue - The default value if field is undefined.
* @return {*} The field value or defaultvalue.
*/
Pikari.getOrDefault = function (field, defaultvalue) {
  let result = Pikari.data.get(field)
  if (result !== undefined) return result
  Pikari.data.set(field, defaultvalue)
  return defaultStatus
}

/**
* @description A helper function to get data field names as an array.
* @return {string[]} Array of field names.
*/
Pikari.getFields = function () {
  return Array.from(Pikari.data.keys())
}

/**
* @description A helper function set wait cursor and disable all elements
* @param {boolean} waiting - true for waiting, false for defaulting cursor and re-enabling elements
*/
Pikari.waiting = function (waiting) {
  if (waiting) {
		document.body.className += (" waiting")
		Pikari._activeElement = document.activeElement
    document.querySelectorAll("body *").forEach(el => { el.disabled = true })
  }
  else {
    const n = document.body.className
    if (n.endsWith(" waiting")) document.body.className = n.substr(0, n.length - 8)
    else document.body.className = n.replace(" waiting ", " ")
		document.querySelectorAll("body *").forEach(el => { el.disabled = false })
		if (Pikari._activeElement) Pikari._activeElement.focus()
  }
}

/**
* @description Acquire locks to data before making changes and committing ("start a transaction").
* <br>Locks prevent concurrent modification by multiple users.
* <br>The basic procedure is to use names-of-fields-to-be-changed as lock names.
* <br>If lock setting fails (because required locks were already locked by other user(s)), all currently held locks by the user are released (to prevent dead-locks).
* <br>Note: Remember to await for the return value, this is an async function.
* @async
* @param {...string} locks - names (or array(s) of names) of the locks to lock. If locks is missing or empty, tries to acquire all existing field names.
* @return {boolean} true if locking was successful. False if user lost all locks and is not eligible to {@link Pikari.commit}.
*/
Pikari.setLocks = async function (...locks) {
  if (!locks || locks.length == 0) locks = Pikari.getFields()
  locks = locks.flat()
  Pikari.locks = "inflight"
	Pikari.waiting(true)
	let response
  try {
    response = await fetch(Pikari._basepath+"setlocks", { method: "post", body: JSON.stringify({ "user": Pikari.user, "pw": Pikari._password, "locks": locks }) })
    if (Pikari.locks === "inflight") {
			if (response.status != 200) {
				Pikari.locks = new Map()
        let w = window.open()
				w.document.write(response.text())
				return false
			}
      Pikari.locks = await response.json()
      if (Pikari.locks["error"]) {
        Pikari._reportError(Pikari.locks["error"])
        Pikari.locks = new Map()
        return false
      }
      Pikari.mylocks = []
      Object.keys(Pikari.locks).forEach(l => { if (Pikari.locks[l].lockedby === Pikari.user) Pikari.mylocks.push(l) })
    }
    if (Pikari.mylocks.length == 0) return false
    Pikari._oldData = new Map()
    Pikari.data.forEach((value, field) => { Pikari._oldData.set(field, JSON.stringify(value)) })
    return true
	}
	catch(e) {
    alert(e.toString())
	} 
  finally {
    Pikari.waiting(false)
  }
}

/**
* @description Commit changes to data fields.
* <br>Only fields that are changed will be transferred to server.
* <br>Will release all locks.
* @throws if no locks are held ("no transaction is active") an error will be thrown
*/
Pikari.commit = function () {
  if (Pikari.mylocks.length == 0) throw ("No transaction to commit")
  let newdata = {}
  Pikari.data.forEach((value, field) => {
    const newvalue = JSON.stringify(value)
    if (!Pikari._oldData.has(field) || Pikari._oldData.get(field) != newvalue) newdata[field] = newvalue
  })
  newdata = JSON.stringify(newdata)
  Pikari._sendToServer("commit", newdata)
}

/**
* @description Rollback any changes to data fields.
* <br>Will cause a local {@link Pikari.changeCallback} with a list of rolled-back (modified) fields.
* @throws if no locks are held ("no transaction is active") an error will be thrown
*/
Pikari.rollback = function () {
  if (Pikari.mylocks.length == 0) throw ("No transaction to rollback")
  let changedfields = []
  Pikari._oldData.forEach((oldvalue, field) => {
    const modifiedvalue = JSON.stringify(Pikari.data.get(field))
    if (oldvalue != modifiedvalue) changedfields.push(field)
    Pikari._oldData.set(field, JSON.parse(oldvalue))
  })
  Pikari.data = Pikari._oldData
  fetch(Pikari._basepath+"setlocks", { method: "post", body: JSON.stringify({ "user": Pikari.user, "pw": Pikari._password, "locks": [] }) })
  for (let l of Pikari._changelisteners) l(Pikari.EVENT.ROLLBACK, changedfields, Pikari.user)
}

/**
* @description Send a message to other on-line users.
* @param message - the message to send
* @param {string|string[]} - a receiver or array of receivers. If missing or empty, the message will be sent to all users.
*/
Pikari.sendMessage = function (message, receivers) {
  if (!receivers) receivers = []
  if (!Array.isArray(receivers)) receivers = [receivers]
  Pikari._ws.send(JSON.stringify({ "sender": Pikari.user, "pw": Pikari._password, "receivers": receivers, "messagetype": "message", "message": message }))
}

/**
* @description Send some string to be added to the server-side log.
* @param {string} event - the string to be added; will be truncated to 10000 letters.
*/
Pikari.log = function (event) {
  if (typeof event !== "string") return
  event = event.substring(0, 10000)
  Pikari._sendToServer("log", event)
}

/**
* @description Log out user. Pikari will respond by closing the web socket connection which triggers {@link Pikari.stopCallback}.
*/
Pikari.logOut = function () {
  Pikari._sendToServer("logout")
}


//private stuff-------------------------------------------

Pikari._stoplistener = null
Pikari._locklisteners = []
Pikari._changelisteners = []
Pikari._messagelisteners = []
Pikari._userlisteners = []

Pikari._app
Pikari._basepath

Pikari._activeElement

Pikari._reportError = function (error) {
  error = "Pikari client error - " + error
  console.log(error)
  if (!error.includes("Web socket problem: ")) Pikari.log(error)
  alert(error)
  throw error
}

Pikari._sendToServer = function (messagetype, message) {
  if (!Pikari._ws) alert("No connection server!")
  else Pikari._ws.send(JSON.stringify({ "sender": Pikari.user, "pw": Pikari._password, "messagetype": messagetype, "message": message }))
}

Pikari._handleStart = function (d) {
  try {
    if (d.message == "wrongpassword") return alert("Wrong password")
    if (d.message == "passwordrequired") return alert("Password required")
    let startdata = JSON.parse(d.message)    
    Object.entries(JSON.parse(startdata.Db)).forEach(([field, data]) => { Pikari.data.set(field, data) })
    const userbag = JSON.parse(startdata.Users)
    let userlist = Object.keys(userbag).sort((a, b) => { return userbag[a] - userbag[b] })
    userlist.forEach((name) => { Pikari.users.set(name, new Date(userbag[name] * 1000)) })
    const changedfields = Pikari.getFields()
    for (let l of Pikari._changelisteners) l(Pikari.EVENT.START, changedfields, "")
    Pikari._startstate = "started"
  }
  finally {
    Pikari.waiting(false)
  }
}

Pikari._handleChange = function (d) {
  const newdata = JSON.parse(d.message)
  if (Object.keys(newdata).length == 0) Pikari.data = new Map()
  Object.entries(newdata).forEach(([field, data]) => {
    if (data == "null") Pikari.data.delete(field); else Pikari.data.set(field, JSON.parse(data))
  })
  for (let l of Pikari._changelisteners) l(Pikari.EVENT.COMMIT, Object.keys(newdata), d.sender)
}

Pikari._handleLocking = function (d) {
  Pikari.locks = JSON.parse(d.message)
  Pikari.mylocks = []
  Object.keys(Pikari.locks).forEach(l => {
    if (Pikari.locks[l].lockedby === Pikari.user) Pikari.mylocks.push(l)
  })
  for (let l of Pikari._locklisteners) l(d.sender)
}

Pikari._handleUser = function (d) {
  if (d.message === Pikari.user) return
  if (d.pw === "in") Pikari.users.set(d.message, new Date())
  else Pikari.users.delete(d.message)
  for (let l of Pikari._userlisteners) l(d.message, d.pw === "in")
}

Pikari._getUrl = function () {
  if (Pikari._app == "index") {
    let url = location.href
    url = url.replace("https://", "wss://")
    url = url.replace("http://", "ws://")
    const end = url.indexOf("?")
    if (end > -1) url = url.substring(0, end)
    return url+"ws?user=" + Pikari.user
  }
  const protocol = location.protocol === "https:" ? "wss://" : "ws://"  
  let host = location.hostname
  if (location.port) host += ":" + location.port
  Pikari._app = location.pathname.substring(1, location.pathname.length)
  if (Pikari._app.endsWith("/")) Pikari._app = Pikari._app.substring(0, Pikari._app.length-1)
  const split = Pikari._app.lastIndexOf('/')
  Pikari._basepath = "/"
  if (split > -1) {
    Pikari._app = Pikari._app.substring(split+1)
    Pikari._basepath = location.pathname.substring(0, split+1) + "/"
  }
  return protocol + host + Pikari._basepath + "ws?user=" + Pikari.user + "&app=" + Pikari._app
}

Pikari._startWebSocket = function () {
	Pikari._ws = new WebSocket(Pikari._getUrl())

  Pikari._ws.onopen = function () {
    Pikari._sendToServer("start")
  }

  Pikari._ws.onclose = function () {
    Pikari._ws = null
    Pikari.data = new Map()
    if (Pikari._startstate != "started") return
    let preventdefault = false
    if (Pikari._stoplistener) preventdefault = Pikari._stoplistener()
    if (!preventdefault) {
      Pikari.waiting(false)
      alert("Connection to Pikari server was lost!")
    }
  }

  Pikari._ws.onmessage = function (evt) {
    const d = JSON.parse(evt.data)
    switch (d.messagetype) {
      case "start": { Pikari._handleStart(d); break }
      case "message": { for (let l of Pikari._messagelisteners) l(d.sender, d.message); break }
      case "lock": { Pikari._handleLocking(d); break }
      case "change": { Pikari._handleChange(d); break }
      case "sign": { Pikari._handleUser(d); break }
      default: Pikari._reportError("Unrecognized message type received: " + d.messagetype)
    }
  }

  Pikari._ws.onerror = function (evt) { Pikari._reportError("Web socket problem: " + evt.data) }
}`
