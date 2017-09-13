package main

import (
	"net/http"
	"io/ioutil"
	"math/rand"
	"path"
	"strings"
	"time"
	"fmt"
	"log"
	"flag"
	"os"
)

const (
	SlideShow = `
<!doctype html>
<html>
<head>
	<title>Slide Show</title>
	<style>
	body, html {
		position: relative;
    	height: 100%;
    	margin: 0;
    	padding: 0;
		margin: 0;
	}

	#container {
    	width: 100%;
    	height: 100%;
    	box-sizing:border-box;
    	background-color: rgb(0,0,0);
	}

	#box {
    	height:100%;
		background-image: none;
    	background-size: contain;
		background-repeat: no-repeat;
		background-position: center;
	}
	img#logo {
		position: absolute;
		bottom: 20px;
		right: 20px;
	}
	</style>
</head>
<body>
	<img src="https://www.suse.com/brandcentral/img/suse/logos_do.png" id="logo" />
	<div id="container">
		<div id="box"></div>
	</div>

	<script type="text/javascript">
		var refresh = function() {
			var box = document.getElementById('box');
			box.style.backgroundImage = "url(" + '/img?' + Math.random() + ")";
    		setTimeout(refresh, 5000);
		};
		refresh();
	</script>
</body>
</html>
`
	UploadForm = `
<!doctype html>
<html>
<head>
	<title>Upload Picture</title>
</head>
<body>
	<p>%s</p>
	<form  method="post" action="/upload" enctype="multipart/form-data">
		<input type="file" name="image"/>
		<input type="submit" value="Upload"/>
	</form>
</body>
</html>
`
)

var imageDir string = "/home/dakechi/images"

func listImageDir() (ret []string) {
	entries, err := ioutil.ReadDir(imageDir)
	if err != nil {
		panic(err)
	}
	ret = make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ret = append(ret, entry.Name())
	}
	return
}

func readRandImage() (contentType string, content []byte) {
	for i := 0; i < 100; i++ {
		entries := listImageDir()
		picked := entries[rand.Intn(len(entries))]
		content, err := ioutil.ReadFile(path.Join(imageDir, picked))
		if err != nil {
			panic(err)
		}
		contentType = http.DetectContentType(content)
		if strings.HasPrefix(contentType, "image/") {
			return contentType, content
		}
		time.Sleep(1 * time.Millisecond)
	}
	return "", []byte{}
}

func main() {
	flag.StringVar(&imageDir, "dir", "./img", "Place to store all images")
	flag.Parse()
	if _, err := os.Stat(imageDir); err != nil {
		panic(err)
	}

	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(SlideShow))
	})

	http.DefaultServeMux.HandleFunc("/img", func(w http.ResponseWriter, r *http.Request) {
		contentType, imageContent := readRandImage()
		w.Header().Set("Content-Type", contentType)
		w.Write(imageContent)
	})

	http.DefaultServeMux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Write([]byte(fmt.Sprintf(UploadForm, "")))
		} else {
			r.ParseMultipartForm(10 * 1048576)
			file, handler, err := r.FormFile("image")
			if err != nil {
				log.Printf("Failed to read form image: %+v", err)
				http.Error(w, "IO error occured during upload", http.StatusBadRequest)
				return
			}
			defer file.Close()
			imageData, err := ioutil.ReadAll(file)
			if err != nil {
				log.Printf("Failed to read image attribute: %+v", err)
				http.Error(w, "Server failed to save your image", http.StatusInternalServerError)
				return
			}
			err = ioutil.WriteFile(path.Join(imageDir, fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)), imageData, 0644)
			if err != nil {
				log.Print(err)
				http.Error(w, "Server failed to save your image", http.StatusInternalServerError)
				return
			}
			w.Write([]byte(fmt.Sprintf(UploadForm, "Successfully uploaded " + handler.Filename)))
		}
	})

	if err := http.ListenAndServe("0.0.0.0:1234", http.DefaultServeMux); err != nil {
		panic(err)
	}
}
