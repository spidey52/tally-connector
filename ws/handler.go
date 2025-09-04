package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// --- HTTP Handler / Upgrader ---

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Adjust this for your origin policy
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	hub.register <- client

	go client.writePump()
	go client.readPump()
}
