package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func createFile(name string, content string, topath ...string) {
	thepath := path
	if len(topath) > 0 {
		thepath = topath[0]
	}
	_, err := os.Stat(thepath + name)
	if os.IsNotExist(err) {
		fmt.Println("Creating " + name + " for " + application)
		file, err := os.Create(thepath + name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			file.WriteString(content)
		}
		file.Close()
	}
}

func createHelloWorld() {
	os.Mkdir(path, 0777)
	createFile("index.html", helloindex)
	//vain kehitys
	//h, _ := ioutil.ReadFile("helloworld.js")
	//createFile("helloworld.js", string(h))
}

func initAssets() {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if application == "HelloWorld" {
			createHelloWorld()
		} else {
			ex, _ := os.Executable()
			fmt.Println("No directory /" + application + " at " + filepath.Dir(ex))
			os.Exit(1)
		}
	}

	if _, err := os.Stat(exePath + "/pikari.js"); os.IsNotExist(err) {
		createFile("pikari.js", pikari, exePath)
	} else {
		p, _ := ioutil.ReadFile("pikari.js")
		pikari = string(p)
	}
}

const helloindex = `
<!DOCTYPE html>
<html lang="en"; style="position: relative; overflow: hidden">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
<title>Pikari - HelloWorld</title>
<noscript>Pikari needs Javascript in order to work.</noscript>
<script src="/pikari.js" type="module"></script>
<script src="/helloworld.js" type="module"></script> 
<style>
body {
	background-color: #FAFAFA;
	text-shadow: 1px 1px 1px rgba(0,0,0,0.010);	
}
body * { box-sizing: border-box; }
</style>
</head>
<body id="body" style="display: flex; margin: 0px; height: 100vh; font-family: system-ui; font-weight: lighter;">
<div id="main">
	<p>Welcome to Pikari!<p>
</div>
</body>
</html>
`

var pikari = `` // const
var helloworld = strings.NewReplacer("¤", "`").Replace(``)
