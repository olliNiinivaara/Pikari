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
Pikari.data = new Map()

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

Pikari.escapeHtml = function(s) {
  s = String(s).replace(/[&<]/g, c => c === '&' ? '&amp;' : '&lt;')
  if (s.startsWith("@")) s = "at" + s.substring(1)
  return s
}

Pikari.start = function(user, password) {
  if (user) Pikari.user = encodeURIComponent(Pikari.escapeHtml(user))
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
  return Array.from(Pikari.data.keys())
}

Pikari.setLocks = async function(locks) {
  if (!locks || (Array.isArray(locks) && locks.length == 0)) locks = Pikari.getFields()
  if (!Array.isArray(locks)) locks  = [String(locks)]
  if (locks.length == 0) throw("No locks to lock")
  Pikari.locks = "inflight"
  let response = await fetch("/setlocks", {method: "post", body: JSON.stringify({"user": "@"+Pikari.user, "password": Pikari.password, "locks": locks})})
  if (Pikari.locks == "inflight") {
    Pikari.locks = await response.json()
    Pikari.mylocks = []
    Object.keys(Pikari.locks).forEach(l => { if (Pikari.locks[l].lockedby == Pikari.userdata) Pikari.mylocks.push(l) })
  }
  if (Pikari.mylocks.length == 0) return false
  Pikari._oldData = new Map()
  Pikari.data.forEach((value, field) => { Pikari._oldData.set(field, JSON.stringify(value)) })
  return true
}

Pikari.commit = function() {
  if (Pikari.mylocks.length == 0) throw("No transaction to commit")
  let newdata = Object.fromEntries(Pikari.data)
  Object.keys(newdata).forEach(k => { newdata[k] = JSON.stringify(newdata[k])})
  newdata = JSON.stringify(newdata)
  Pikari._sendToServer("commit", newdata)
}

Pikari.rollback = function() {
  //TODO: report only rollbacked fields, NOT all fields
  if (Pikari.mylocks.length == 0) throw("No transaction to rollback")
  Pikari.data = new Map()
  Pikari._oldData.forEach(v, f => { Pikari.data.set(f, JSON.parse(v)) })
  const allfields = Pikari.getFields()
  for(let l of Pikari._changelisteners) l("rollback", allfields)
  fetch("/setlocks", {method: "post", body: JSON.stringify({"user": "@"+Pikari.user, "password": Pikari.password, "locks": []})})
}

Pikari.dropData = function() {
  Pikari._sendToServer("dropdata")
}

Pikari.setUserdata = function(value) {
  Pikari.data.set("@"+Pikari.user, value)
}

Pikari.getUserdata = function() {
  return Pikari.data.get("@"+Pikari.user)
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
  else Pikari._ws.send(JSON.stringify({"sender": "@"+Pikari.user, "password":Pikari._password, "messagetype": messagetype, "message": message}))
}

Pikari._handleStart = function(d) {
  const startdata = JSON.parse(d.message)
  Object.entries(startdata).forEach(([field, data]) => { Pikari.data.set(field, data) })
  for(let l of Pikari._startlisteners) l()  
}

Pikari._handleChange = function(d) {
  const newdata = JSON.parse(d.message)
  if (Object.keys(newdata).length == 0) Pikari.data = new Map()
  Object.entries(newdata).forEach(([field, data]) => { if (data == "null") Pikari.data.delete(field); else Pikari.data.set(field, JSON.parse(data)) })
  for(let l of Pikari._changelisteners) l(d.sender, Object.keys(newdata))
}

Pikari._handleLocking = function(d) {
  const lockmessage = JSON.parse(d.message)
  Pikari.locks = lockmessage.locks
  Pikari.mylocks = []
  Object.keys(Pikari.locks).forEach(l => {
    Pikari.locks[l].lockedby = Pikari.locks[l].lockedby.substring(1)
    if (Pikari.locks[l].lockedby == Pikari.user) Pikari.mylocks.push(l)
  })
  for(let l of Pikari._locklisteners) l(d.sender, lockmessage.incremented)
}

Pikari._startWebSocket = function() {
  Pikari._ws = new WebSocket("ws://"+document.location.host+"/ws?user="+"@"+Pikari.user)

  Pikari._ws.onopen = function() {  
    Pikari._sendToServer("start")
  }

  Pikari._ws.onclose = function() {
    Pikari._ws = null
    Pikari.data = new Map()
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
