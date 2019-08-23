package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type transactionrequest struct {
	User   string   `json:"user"`
	Fields []string `json:"fields"`
}

type lockedfield struct {
	//Name: string `json:"field"`
	lockedby    string //`json:"lockedby"`
	lockedsince string
}

var lockedfields = make(map[string]*lockedfield)

func startTransaction(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request transactionrequest
	err := decoder.Decode(&request)
	if err != nil {
		log.Println("Pikari server error - transaction request parsing error: " + err.Error())
		return
	}
	fmt.Println("tuli")
	fmt.Println(request)
	w.Header().Set("Content-Type", "text/json")
	fmt.Fprint(w, "[]")
}
