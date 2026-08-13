package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	dir := publicDir()
	// tcp4 so port-forwarders that only scan IPv4 sockets can see :8080.
	ln, err := net.Listen("tcp4", "0.0.0.0:8080")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("serving %s on http://localhost:8080", dir)
	log.Fatal(http.Serve(ln, http.FileServer(http.Dir(dir))))
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
