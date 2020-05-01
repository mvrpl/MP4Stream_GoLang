package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

const playlistM3U8 = "stream.m3u8"

var ffmpegProc int

func main() {
	host := flag.String("h", "", "Host m3u8 streaming (Opcional)")
	port := flag.Int("p", 8100, "Porta m3u8 streaming (Opcional)")
	video := flag.String("v", "", "Caminho arquivo MP4 (Obrigatorio)")
	flag.Parse()

	RemoveContents()

	if isMP4(*video) {
		flag.PrintDefaults()
		os.Exit(1)
	}

	go playList(*video)

	serverFolder := filepath.Join(filepath.Dir(os.Args[0]), "stream")

	handler := addHeaders(http.FileServer(http.Dir(serverFolder)))

	http.Handle("/", handler)

	s := &http.Server{
		Addr:           *host + ":" + strconv.Itoa(*port),
		Handler:        handler,
		ReadTimeout:    3 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err := s.ListenAndServe()

	if err != nil {
		fmt.Println(err)
		runtime.Goexit()
		os.Exit(1)
	}
}

func isMP4(file string) bool {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	if mime := mimetype.Detect(data); mime.Is("audio/mp4") {
		return true
	} else {
		return false
	}
}

func RemoveContents() {
	os.RemoveAll(filepath.Join(filepath.Dir(os.Args[0]), "stream"))
	os.MkdirAll(filepath.Join(filepath.Dir(os.Args[0]), "stream"), 0777)
}

func playList(video string) {
	filePath := filepath.Join(filepath.Dir(os.Args[0]), "stream", playlistM3U8)
	cmdArguments := []string{
		"-threads", "2", "-re", "-fflags", "+genpts", "-stream_loop", "-1", "-i", video, "-c", "copy", filePath,
	}
	cmd := exec.Command("ffmpeg", cmdArguments...)
	cmd.Start()
}

func addHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	}
}
