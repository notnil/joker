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
	h := http.FileServer(http.Dir(dir))

	// Listen on IPv4 and IPv6 separately. A dual-stack :8080 socket only
	// shows up in the IPv6 table (port-forwarders miss it). tcp4-only
	// refuses Chrome's localhost, which resolves to ::1 first.
	ln4, err := net.Listen("tcp4", "0.0.0.0:8080")
	if err != nil {
		log.Fatal(err)
	}
	ln6, err := net.Listen("tcp6", "[::]:8080")
	if err != nil {
		log.Printf("IPv6 listen skipped: %v", err)
		log.Printf("serving %s on http://127.0.0.1:8080", dir)
		log.Fatal(http.Serve(ln4, h))
	}
	log.Printf("serving %s on http://localhost:8080", dir)
	go func() { log.Fatal(http.Serve(ln6, h)) }()
	log.Fatal(http.Serve(ln4, h))
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
