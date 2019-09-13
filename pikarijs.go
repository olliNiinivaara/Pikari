package main

const pikari = `
/**
 * @file Pikari API
 * @see https://github.com/olliNiinivaara/Pikari
 * @author Olli Niinivaara
 * @copyright Olli Niinivaara 2019
 * @license MIT
 * @version 0.5
 */


/** @namespace
 * @description the global Pikari object. To initialize, add listeners and call start.
 * @global
 * @example Pikari.addCommitListener(function(description, sender, fields) { do something })
 * Pikari.start("John")
*/
window.Pikari = new Object()


/** 
 *  @description Local copy of the whole database. Changes committed to database by any user are automatically updated to it.
 *  @example Pikari.data[Pikari.userdata] = "some data"
 */
Pikari.data = {}

/** 
 * @description Names of locks that are currently held by the current user.
 * At least one lock must be held before commit can be called.
 */
Pikari.mylocks = []


/**
 @typedef Pikari.Lock
 @type {Object}
 @property {string} lockedby {@link Pikari.userdata} identifier of current lock owner 
 @property {string} lockedsince The start time of locking
 */
/** 
 * @description Locks that are held.
 * Object's property names (keys) identify the locks and properties are of type {@link Pikari.Lock}
 * @example console.log(Object.keys(Pikari.mylocks))
 */
Pikari.locks = {}

Pikari.generateKey = function() {
  return Math.floor(Math.random() * Number.MAX_SAFE_INTEGER)
}


/** 
 * @description Name of the user.
 */
Pikari.user = "Anon-"+ Pikari.generateKey()


/** 
 * @description Name of the user's personal database field (simply user with a leading @ - to prevent name clashes)
 */
Pikari.userdata = "@" + Pikari.user


/**
 * Helper function to strip a leading character from string
 * If user can set id's like user names, database fields or locks,
 * they should be escaped with an extra character to prevent name clashes with other ids
 * (and javascript object prototype properties)
 * @param {string} escapedstring - (starting with an extra character)
 * @returns {string} original string
 */
Pikari.unescId = function(escapedstring) {
  return escapedstring.substring(1)
}

Pikari.escapeHtml = function(s) {
  return String(s).replace(/[&<]/g, c => c === '&' ? '&amp;' : '&lt;')
}

Pikari.start = function(user, password) {
  if (user) {
    Pikari.user = encodeURIComponent(escapeHtml(user))
    Pikari.userdata = "@" + Pikari.user
  }
  if (!password) password = ""
  Pikari._password = String(password)
  Pikari._startWebSocket()
}

Pikari.addStartListener = function(handler) {
  Pikari._startlisteners.push(handler)
}

Pikari.addStopListener = function(handler) {
  Pikari._stoplisteners.push(handler)
}

Pikari.addChangeListener = function(handler) {
  Pikari._changelisteners.push(handler)
}

Pikari.addMessageListener = function(handler) {
  Pikari._messagelisteners.push(handler)
}

Pikari.addLockListener = function(handler) {
  Pikari._locklisteners.push(handler)
}

Pikari.log = function(event) {
  Pikari._sendToServer("log", event)
}

Pikari.sendMessage = function(message, receivers) {
  if (!receivers) receivers = []
  Pikari._sendToServer("message", JSON.stringify({"receivers": receivers, "message": message}))
}

Pikari.getFields = function() {
  return Object.keys(Pikari.data)
}

Pikari.setLocks = async function(locks) {
  if (!locks || (Array.isArray(locks) && locks.length == 0)) locks = Pikari.getFields()
  if (!Array.isArray(locks)) locks  = [String(locks)]
  if (locks.length == 0) throw("No locks to lock")
  let response = await fetch("/setlocks", {method: "post", body: JSON.stringify({"user": Pikari.userdata, "password": Pikari.password, "locks": locks})})
  Pikari.locks = await response.json()
  Pikari.mylocks = []
  Object.keys(Pikari.locks).forEach(l => { if (Pikari.locks[l].lockedby == Pikari.userdata) Pikari.mylocks.push(l) })
  if (Pikari.mylocks.length == 0) return false
  Pikari._oldData = {}
  Pikari.getFields().forEach(f => { Pikari._oldData[f] = JSON.stringify(Pikari.data[f]) })
  return true
}

Pikari.commit = function() {
  if (Pikari.mylocks.length == 0) throw("No transaction to commit")
  let newdata = {}
  Pikari.getFields().forEach(f => {
    const value = JSON.stringify(Pikari.data[f])
    if (!Pikari._oldData[f] || Pikari._oldData[f] != value) newdata[f] = value
  })
  newdata = JSON.stringify(newdata)
  Pikari._sendToServer("commit", newdata)
}

Pikari.rollback = async function() {
  if (Pikari.mylocks.length == 0) throw("No transaction to rollback")
  const r = await fetch("/setlocks", {method: "post", body: JSON.stringify({"user": Pikari.userdata, "password": Pikari.password, "locks": []})})
  await r.text()
  Pikari.data = JSON.parse(JSON.stringify(Pikari._oldData))
}

Pikari.dropData = function() {
  Pikari._sendToServer("dropdata")
}


//private stuff-------------------------------------------

Pikari._startlisteners = []
Pikari._stoplisteners = []
Pikari._locklisteners = []
Pikari._changelisteners = []
Pikari._messagelisteners = []

Pikari._reportError = function(error) {
  error = "Pikari client error - " + error
  console.log(error)
  if (!error.includes("Web socket problem: ")) Pikari.log(error)
  alert(error)
  throw error
}

Pikari._sendToServer = function(messagetype, message) {
  if (!Pikari._ws) alert ("No connection server!")
  else Pikari._ws.send(JSON.stringify({"sender": Pikari.userdata, "password":Pikari._password, "messagetype": messagetype, "message": message}))
}

Pikari._handleStart = function(d) {
  const startdata = JSON.parse(d.message)
  Object.entries(startdata).forEach(([field, data]) => { Pikari.data[field] = data })
  for(let l of Pikari._startlisteners) l()  
}

Pikari._handleChange = function(d) {
  const newdata = JSON.parse(d.message)
  if (Object.keys(newdata).length == 0) Pikari.data = {}
  Object.entries(newdata).forEach(([field, data]) => {
    Pikari.data[field] = JSON.parse(data)
    if (Pikari.data[field] == null) delete Pikari.data[field]
  })
  for(let l of Pikari._changelisteners) l(d.sender, Object.keys(newdata))
}

Pikari._handleLocking = function(d) {
  const lockmessage = JSON.parse(d.message)
  Pikari.locks = lockmessage.locks
  Pikari.mylocks = []
  Object.keys(Pikari.locks).forEach(l => { if (Pikari.locks[l].lockedby == Pikari.userdata) Pikari.mylocks.push(l) })
  for(let l of Pikari._locklisteners) l(d.sender, lockmessage.incremented)
}

Pikari._startWebSocket = function() {
  Pikari._ws = new WebSocket("ws://"+document.location.host+"/ws?user="+Pikari.userdata)

  Pikari._ws.onopen = function() {  
    Pikari._sendToServer("start")
  }

  Pikari._ws.onclose = function() {
    Pikari._ws = null
    Pikari.data = {}
    for(let l of Pikari._stoplisteners) l()
  }

  Pikari._ws.onmessage = function(evt) {
    const d = JSON.parse(evt.data)
    switch (d.messagetype) {
      case "start": { Pikari._handleStart(d); break }
      case "message": { for(let l of Pikari._messagelisteners) l(d.sender, d.message); break }
      case "lock":  { Pikari._handleLocking(d); break }      
      case "change": { Pikari._handleChange(d); break }
      default: Pikari._reportError("Unrecognized message type received: " + d.messagetype)
    }
  }

  Pikari._ws.onerror = function(evt) { Pikari._reportError("Web socket problem: " + evt.data) }
}
`
