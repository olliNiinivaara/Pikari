package main

const pikari = `
/**
 * @file Pikari API
 * @see https://github.com/olliNiinivaara/Pikari
 * @author Olli Niinivaara
 * @copyright Olli Niinivaara 2019
 * @license MIT
 * @version 0.4
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
 * Helper function to strip the leading @
 * @param {string} userdata - A user identifier (starting with @)
 * @returns {string} User's name
 */
Pikari.getUsername = function(userdata) {
  return userdata.substring(1)
}

Pikari.escape = function(s) {
  return String(s).replace(/[&<]/g, c => c === '&' ? '&amp;' : '&lt;')
}

Pikari.start = function(user, password) {
  if (user) {
    Pikari.user = encodeURIComponent(escape(user))
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

Pikari.getFields = function() {
  return Object.getOwnPropertyNames(Pikari.data)
}

Pikari.startTransaction = function(fields) {
  if (Pikari._fieldsInThisTransaction) throw("Transaction is already active")
  if (!fields || (Array.isArray(fields) && fields.length == 0)) fields = Pikari.getFields()
  if (!Array.isArray(fields)) fields = [String(fields)]
  if (fields.length == 0) throw("No fields to transact with")
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

Pikari.drop = function() {
  Pikari._sendToServer("drop")
}


//private stuff-------------------------------------------

Pikari._startlisteners = []
Pikari._stoplisteners = []
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
    for(let l of Pikari._startlisteners) l()
  } else {
    Object.entries(dd.fields).forEach(([field, data]) => {
      Pikari.data[field] = JSON.parse(data)
      if (Pikari.data[field] == null) delete Pikari.data[field]
      delete Pikari.fieldsInTransaction[field]
    })
    if (d.sender == Pikari.userdata) Pikari._fieldsInThisTransaction = false
    for(let l of Pikari._commitlisteners) l(d.sender, Object.keys(dd), dd.description)
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

Pikari._handleDrop = function() {
  Pikari.fieldsInTransaction = {}
  Pikari._fieldsInThisTransaction = false
  Pikari.data = {}
  for(let l of Pikari._startlisteners) l("drop")
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
      case "message": { for(let l of Pikari._messagelisteners) l(d.sender, d.message); break }
      case "transaction":  { Pikari._handleTransaction(d); break }      
      case "change": { Pikari._handleCommit(d); break }
      case "rollback":  { Pikari._handleRollback(d); break }
      case "drop":  { Pikari._handleDrop(); break }
      default: Pikari._reportError("Unrecognized message type received: " + d.messagetype)
    }
  }

  Pikari._ws.onerror = function(evt) { Pikari._reportError("Web socket problem: " + evt.data) }
}
`
