package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	os.Mkdir(exedir+dir, 0700)
	files := formdata.File["files"]
	for i := range files {
		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			fmt.Fprintln(w, err)
			return
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
			return
		}
		_, err = io.Copy(out, file)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
	}
}

func updateAdmindata(dir, value string) {
	tx, err := apps["admin"].database.Begin()
	if err != nil {
		log.Fatal("Pikari server error - could not start transaction: " + err.Error())
	}
	update(apps["admin"], tx, dir, value)
	if err = tx.Commit(); err != nil {
		log.Fatal("Pikari server error - could not commit data: " + err.Error())
	}
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
