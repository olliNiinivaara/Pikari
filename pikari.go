package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
	"unicode/utf8"

	"gopkg.in/natefinch/lumberjack.v2"
)

var exedir, port = "", 8080

func main() {
	rand.Seed(time.Now().UnixNano())
	icon, _ = base64.StdEncoding.DecodeString(icon64)
	exedir, _ = os.Getwd()
	exedir += string(filepath.Separator)
	var pw string
	var port int
	flag.StringVar(&pw, "password", "", "password for the application")
	flag.IntVar(&port, "port", 8080, "IP port")
	flag.Parse()
	if pw == "" {
		pw = generatePassword()
	}
	pw = base64.StdEncoding.EncodeToString([]byte(pw))
	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.SetOutput(&lumberjack.Logger{
		Filename:   exedir + "pikari.log",
		MaxSize:    1,
		MaxBackups: 3,
		LocalTime:  true,
	})

	fmt.Println("Pikari 0.9 starting at " + exedir)
	addr := "127.0.0.1:" + strconv.Itoa(port)
	initApps(pw)
	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/ws", ws)
	http.HandleFunc("/setlocks", setLocks)
	http.HandleFunc("/admin/dirupload", dirUploadHandler)
	http.HandleFunc("/admin/gitupload", gitUploadHandler)
	http.HandleFunc("/admin/update", updateHandler)
	http.HandleFunc("/admin/delete/", deleteHandler)
	createPikariJs()
	rootfs := http.FileServer(http.Dir(exedir))
	http.Handle("/", rootfs)
	fmt.Println("Serving to " + addr)
	fmt.Println("Send SIGINT (Ctrl+C) to quit")
	log.Println("---STARTED---")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		closeDbs()
		os.Exit(0)
	}()
	err := http.ListenAndServe(addr, nil)
	fmt.Println(err)
	log.Fatal(err)
}

func createPikariJs() {
	if _, err := os.Stat(exedir + "pikari.js"); err == nil {
		return
	}
	file, err := os.Create(exedir + "pikari.js")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	file.WriteString(pikari)
	file.Close()
}

func generatePassword() string {
	letters := []rune("abcdefghijkmnpqrstuvxyzABCDEFGHIJKLMNPQRSTUVXYZ23456789")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	fmt.Println("Admin password: " + string(b))
	return string(b)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}
