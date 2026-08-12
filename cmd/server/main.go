package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	dir := publicDir()
	log.Printf("serving %s on :8080", dir)
	log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir(dir))))
}

func publicDir() string {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		p := filepath.Join(filepath.Dir(file), "public")
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	return "public"
}
