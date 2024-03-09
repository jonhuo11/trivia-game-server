package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func boolPtr(v bool) *bool {
	x := v
	return &x
}

var addr = flag.String("addr", ":9100", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	io.WriteString(w, "Welcome")
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	player := &Player{
		id:   fmt.Sprintf("Player %v", uuid.New().String()),
		hub:  hub,
		conn: conn,
		send: make(chan OutgoingMessage, 512), // buffer the send channel by 512 messages to prevent panic overflow
	}
	player.hub.register <- player

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go player.writePump()
	go player.readPump()
}

func main() {
	fmt.Println("Starting server")
	flag.Parse()
	hub := newHub()
	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	fmt.Println("Serving on ", addr)
}
