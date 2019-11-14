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
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"gopkg.in/natefinch/lumberjack.v2"
)

const s = string(filepath.Separator)

var exedir, port = "", 8080

const icon64 = "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"

var icon []byte

func favicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=7776123")
	w.Write(icon)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	icon, _ = base64.StdEncoding.DecodeString(icon64)
	exedir, _ = os.Getwd()
	exedir += s
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
	createFiles()
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

func createFiles() {
	if _, err := os.Stat(exedir + "pikari.js"); err != nil {
		file, err := os.Create(exedir + "pikari.js")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		file.WriteString(pikari)
		file.Close()
	}
	if _, err := os.Stat(exedir + "index.html"); err != nil {
		file, err := os.Create(exedir + "index.html")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		file.WriteString(index1 + "`" + index2 + "`" + index3)
		file.Close()
	}
	if _, err := os.Stat(exedir + "admin" + s + "admin.mjs"); err != nil {
		file, err := os.Create(exedir + "admin" + s + "admin.mjs")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		file.WriteString(strings.Replace(adminmjs, "¤", "`", -1))
		file.Close()
	}
	if _, err := os.Stat(exedir + "admin" + s + "index.html"); err != nil {
		file, err := os.Create(exedir + "admin" + s + "index.html")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		file.WriteString(adminindex)
		file.Close()
	}
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
