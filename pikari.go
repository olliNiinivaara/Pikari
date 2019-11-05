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

// const tf = "01-02 15:04"

var exedir, port = "", 8080

var config configuration

type configuration struct {
	Port         int
	Maxpagecount int
	Autorestart  bool
	Usercount    bool
}

func main() {
	rand.Seed(time.Now().UnixNano())
	icon, _ = base64.StdEncoding.DecodeString(icon64)
	exedir, _ = os.Getwd()
	exedir += string(filepath.Separator)
	// flag.StringVar(&appdir, "appdir", "", "path to application, absolute or relative to "+exedir)
	var pw string
	var port int
	flag.StringVar(&pw, "password", "", "password for the application")
	flag.IntVar(&port, "port", 8080, "IP port")
	flag.Parse()
	if pw == "" {
		pw = generatePassword()
	}
	pw = base64.StdEncoding.EncodeToString([]byte(pw))

	/*if len(appdir) == 0 {
		fmt.Println("Give path to application with appdir parameter, like this: pikari -appdir Nameofmyapplication")
		os.Exit(1)
	}
	if !strings.HasSuffix(appdir, string(filepath.Separator)) {
		appdir += string(filepath.Separator)
	}
	if !filepath.IsAbs(appdir) {
		appdir = exedir + appdir
	}
	_, err := os.Stat(appdir)
	if os.IsNotExist(err) {
		fmt.Println("Application directory not found: " + appdir)
		os.Exit(1)
	}*/
	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.SetOutput(&lumberjack.Logger{
		Filename:   exedir + "pikari.log",
		MaxSize:    1,
		MaxBackups: 3,
		LocalTime:  true,
	})

	fmt.Println("Pikari 0.8 starting at " + exedir)
	//readConfig()
	addr := "127.0.0.1:" + strconv.Itoa(port)
	initApps(pw)
	//openDb(config.Maxpagecount)
	//getData()
	//fs := http.FileServer(http.Dir(exedir))
	//http.Handle("/", fs)
	//http.Handle("/index/", http.StripPrefix("/index/", portalfs))
	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/ws", ws)
	http.HandleFunc("/setlocks", setLocks)
	createPikariJs()
	rootfs := http.FileServer(http.Dir(exedir))
	http.Handle("/", rootfs)
	fmt.Println("Serving to " + addr)
	fmt.Println("Send SIGINT (Ctrl+C) to quit")
	/*if config.Usercount {
		fmt.Print(time.Now().Format(tf) + " users: 0" + " ")
	}*/
	log.Println("---")

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

/*func readConfig() {
	createPikariToml()
	if _, err := toml.DecodeFile(exedir+"pikari.toml", &config); err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}*/

/*func createAdmin() *appstruct {
	os.Mkdir(exedir+"admin", 0700)
	return createAdminApp()
}*/

/*func createPikariToml() {
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
}*/

func generatePassword() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
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
