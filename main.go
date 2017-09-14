package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
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
    	height: 100%%;
    	margin: 0;
    	padding: 0;
		margin: 0;
	}

	#container {
    	width: 100%%;
    	height: 100%%;
    	box-sizing:border-box;
    	background-color: rgb(0,0,0);
	}

	#box {
    	height:100%%;
		background-image: none;
    	background-size: contain;
		background-repeat: no-repeat;
		background-position: center;
	}
	img#logo {
		position: absolute;
		bottom: 20px;
		right: 20px;
		opacity: .7;
	}
	</style>
</head>
<body>
	<img src="/logo" id="logo" />
	<div id="container">
		<div id="box"></div>
	</div>

	<script type="text/javascript">
		var refresh = function() {
			var box = document.getElementById('box');
			box.style.backgroundImage = "url(" + '/img?' + Math.random() + ")";
    		setTimeout(refresh, %d);
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
var logoPath string = "/home/dakechi/logo.png"
var intervalMS int = 5000

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
	flag.StringVar(&imageDir, "dir", "./img", "Directory of slide show images (jpg, png, gif, etc)")
	flag.StringVar(&logoPath, "logo", "./logo", "Logo to display at bottom right corner during slide show")
	flag.IntVar(&intervalMS, "intervalms", 5000, "Slide show speed (interval in milliseconds)")
	flag.Parse()
	if _, err := os.Stat(imageDir); err != nil {
		panic(err)
	}

	logoData, err := ioutil.ReadFile(logoPath)
	if err != nil {
		panic(err)
	}
	logoDataContentType := http.DetectContentType(logoData)
	if !strings.HasPrefix(logoDataContentType, "image") {
		panic("The logo image does not look like an image")
	}

	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Print(err)
			http.Error(w, "Sorry, your access is not authorized", http.StatusForbidden)
			return
		}

		//We allow slideshow access only to localhost to avoid people copying Photos inside party network.
		if host == "::1" || host == "127.0.0.1" {
			w.Write([]byte(fmt.Sprintf(SlideShow, intervalMS)))

		} else {
			log.Printf(r.RemoteAddr)
			http.Error(w, "Sorry, your access is not authorized", http.StatusForbidden)
			return
		}
	})

	http.DefaultServeMux.HandleFunc("/img", func(w http.ResponseWriter, r *http.Request) {
		contentType, imageContent := readRandImage()
		w.Header().Set("Content-Type", contentType)
		w.Write(imageContent)
	})

	http.DefaultServeMux.HandleFunc("/logo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", logoDataContentType)
		w.Write(logoData)
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
			w.Write([]byte(fmt.Sprintf(UploadForm, "Successfully uploaded "+handler.Filename)))
		}
	})

	if err := http.ListenAndServe("0.0.0.0:1234", http.DefaultServeMux); err != nil {
		panic(err)
	}
}
