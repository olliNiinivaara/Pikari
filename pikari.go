package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/natefinch/lumberjack.v2"
)

const tf = "01-02 15:04"

var appdir, exedir, port, password = "", "", 0, ""

var config configuration

type configuration struct {
	Port         int
	Maxpagecount int
	Autorestart  bool
}

func main() {
	icon, _ = base64.StdEncoding.DecodeString(icon64)
	_, callerFile, _, _ := runtime.Caller(0)
	exedir = filepath.Dir(callerFile) + string(filepath.Separator)
	flag.StringVar(&appdir, "appdir", "", "path to application, absolute or relative to "+exedir)
	var pw string
	flag.StringVar(&pw, "password", "", "password for the application")
	flag.Parse()
	if len(appdir) == 0 {
		fmt.Println("Give path to application with appdir parameter, like this: pikari -appdir Nameofmyapplication")
		os.Exit(1)
	}
	password = base64.StdEncoding.EncodeToString([]byte(pw))
	if !filepath.IsAbs(appdir) {
		appdir = exedir + appdir + string(filepath.Separator)
	}
	_, err := os.Stat(appdir)
	if os.IsNotExist(err) {
		fmt.Println("Application directory not found: " + appdir)
		os.Exit(1)
	}
	log.SetOutput(&lumberjack.Logger{
		Filename:   appdir + "pikari.log",
		MaxSize:    1,
		MaxBackups: 3,
		LocalTime:  true,
	})
	fmt.Println(time.Now().Format(tf) + " Pikari 0.8 starting")
	readConfig()
	addr := "127.0.0.1:" + strconv.Itoa(config.Port)
	openDb(config.Maxpagecount)
	getData()
	fs := http.FileServer(http.Dir(appdir))
	http.Handle("/", fs)
	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/ws", ws)
	http.HandleFunc("/setlocks", setLocks)
	createPikariJs()
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
	_, err := os.Stat(exedir + "pikari.js")
	if err == nil {
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

func readConfig() {
	createPikariToml()
	if _, err := toml.DecodeFile("pikari.toml", &config); err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}

func createPikariToml() {
	if _, err := os.Stat(exedir + "pikari.toml"); err == nil {
		return
	}
	file, err := os.Create(exedir + "pikari.toml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	file.WriteString(tomlconfig)
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
