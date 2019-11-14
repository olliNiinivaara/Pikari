package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func dirUploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(200000)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	formdata := r.MultipartForm
	defer formdata.RemoveAll()
	dir := formdata.Value["dir"][0]
	if dir == "" {
		fmt.Fprintln(w, "Directory value not given")
		return
	}
	if dir[0] == filepath.Separator {
		dir = dir[1 : len(dir)-1]
	}
	if _, err := os.Stat(exedir + dir); !os.IsNotExist(err) {
		fmt.Fprintln(w, "An application already exists at directory "+exedir+dir)
		return
	}
	if !copyFiles(dir, formdata.File["files"], w) {
		return
	}
	apps[dir] = &appstruct{Name: dir, Maxpagecount: -1, Autorestart: -1, Source: formdata.Value["source"][0], Git: "", exists: true}
	b, _ := json.Marshal(apps[dir])
	updateAdmindata(dir, string(b))
}

func gitUploadHandler(w http.ResponseWriter, r *http.Request) {
	dir := r.FormValue("dir")
	if dir == "" {
		fmt.Fprintln(w, "Directory value not given")
		return
	}
	if dir[0] == filepath.Separator {
		dir = dir[1 : len(dir)-1]
	}
	if _, err := os.Stat(exedir + dir); !os.IsNotExist(err) {
		fmt.Fprintln(w, "An application already exists at directory "+exedir+dir)
		return
	}
	u := r.FormValue("url")
	url, err := url.ParseRequestURI(u)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	if !cloneRepo(dir, url, w) {
		return
	}
	apps[dir] = &appstruct{Name: dir, Maxpagecount: -1, Autorestart: -1, Source: url.String(), Git: "1", exists: true}
	b, _ := json.Marshal(apps[dir])
	updateAdmindata(dir, string(b))
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(200000)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	formdata := r.MultipartForm
	defer formdata.RemoveAll()
	dir := formdata.Value["dir"][0]
	appmutex.Lock()
	defer appmutex.Unlock()
	if dir == "" || apps[dir] == nil {
		fmt.Fprintln(w, "App does not exist: "+dir)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	disabled := 0
	if formdata.Value["disabled"] != nil {
		disabled = 1
	}
	disabledchanged := disabled != apps[dir].Disabled
	apps[dir].Disabled = disabled
	sourcechanged := formdata.Value["source"][0] != apps[dir].Source
	if disabledchanged || sourcechanged {
		apps[dir].Source = formdata.Value["source"][0]
		b, _ := json.Marshal(apps[dir])
		updateAdmindata(dir, string(b))
	}
	closeApp(dir)
	deletedata := formdata.Value["deletedata"]
	datafile := exedir + dir + string(filepath.Separator) + "data.db"
	var files []*multipart.FileHeader
	var giturl *url.URL
	if formdata.Value["dogit"] == nil {
		files = formdata.File["files"]
	} else {
		if formdata.Value["source"][0] == "" {
			fmt.Fprintln(w, "Repo URL missing, cannot update")
			return
		}
		giturl, err = url.ParseRequestURI(formdata.Value["source"][0])
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
	}
	if formdata.Value["dogit"] == nil && (files == nil || len(files) == 0) {
		if deletedata != nil {
			os.Remove(datafile)
		}
		return
	}
	var data []byte
	if deletedata == nil {
		data, _ = ioutil.ReadFile(datafile)
	}
	os.RemoveAll(exedir + dir)
	if files != nil {
		copyFiles(dir, files, w)
	} else {
		cloneRepo(dir, giturl, w)
	}
	if data != nil {
		err = ioutil.WriteFile(datafile, data, 0644)
		if err != nil {
			fmt.Fprintln(w, "Error preserving data.db for app "+dir)
		}
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	app := r.URL.Query().Get("app")
	appmutex.Lock()
	defer appmutex.Unlock()
	if app == "" || apps[app] == nil {
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	closeApp(app)
	updateAdmindata(app, "null")
	apps[app] = nil
	os.RemoveAll(exedir + app)
}

func updateAdmindata(dir, value string) {
	admin := apps["admin"]
	tx, err := admin.database.Begin()
	if err != nil {
		log.Fatal("Pikari server error - could not start transaction: " + err.Error())
	}
	update(admin, tx, dir, value)
	if err = tx.Commit(); err != nil {
		log.Fatal("Pikari server error - could not commit data: " + err.Error())
	}
	admin.buffer.Reset()
	s, _ := json.Marshal(value)
	serialized := `{"` + dir + `":` + string(s) + `}`
	transmitMessage(apps["admin"], &wsdata{Sender: "admin", Receivers: []string{}, Messagetype: "change", Message: serialized}, false)
}

func updateApp(dir, value string) {
	appmutex.Lock()
	defer appmutex.Unlock()
	if err := json.Unmarshal([]byte(value), apps[dir]); err != nil {
		log.Fatal(err)
	}
	if dir != "admin" {
		closeApp(dir)
	} else {
		for _, app := range apps {
			if app.Maxpagecount == -1 {
				closeApp(dir)
			}
		}
	}
}

func copyFiles(dir string, files []*multipart.FileHeader, w http.ResponseWriter) bool {
	if len(files) == 0 {
		return true
	}
	os.Mkdir(exedir+dir, 0700)
	for i := range files {
		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			fmt.Fprintln(w, err)
			return false
		}
		name := files[i].Filename
		index := strings.Index(name, string(filepath.Separator))
		if index != -1 {
			name = name[index:]
		}
		os.MkdirAll(exedir+dir+filepath.Dir(name), 0700)
		out, err := os.Create(exedir + dir + name)
		defer out.Close()
		if err != nil {
			fmt.Fprintf(w, "Unable to create file for writing: "+files[i].Filename)
			return false
		}
		_, err = io.Copy(out, file)
		if err != nil {
			fmt.Fprintln(w, err)
			return false
		}
	}
	return true
}

func cloneRepo(dir string, url *url.URL, w http.ResponseWriter) bool {
	os.Mkdir(exedir+dir, 0700)
	cmd := exec.Command("git", "clone", "--depth", "1", url.String(), ".")
	cmd.Dir = exedir + dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(exedir + dir)
		fmt.Fprintln(w, string(out)+" "+err.Error())
		return false
	}
	os.RemoveAll(exedir + dir + string(filepath.Separator) + ".git")
	return true
}
