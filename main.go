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
	<img src="%s" id="logo" />
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
	LogoSrc = `data:image/png;base64,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`
)

var imageDir string = "/home/dakechi/images"
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
	flag.IntVar(&intervalMS, "intervalms", 5000, "Slide show speed (interval in milliseconds)")
	flag.Parse()
	if _, err := os.Stat(imageDir); err != nil {
		panic(err)
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
			w.Write([]byte(fmt.Sprintf(SlideShow, LogoSrc, intervalMS)))

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
