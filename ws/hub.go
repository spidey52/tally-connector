package ws

import (
	"encoding/json"
	"log"
	"sync"
)

// --- Hub ---

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex // Added a mutex to protect the clients map
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Println("Client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				client.conn.Close()
				log.Println("Client disconnected")
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
					client.conn.Close()
				}
			}
			h.mu.Unlock()
		}
	}
}

// GetClientCount returns the current number of connected clients.
func (h *Hub) GetClientCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

func (h *Hub) Broadcast(message Message) {
	// Convert the Message struct to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshalling message:", err)
		return
	}

	h.broadcast <- messageJSON
}

// GetClients returns a safe copy of the list of connected clients.
func (h *Hub) GetClients() []*Client {
	h.mu.Lock()
	defer h.mu.Unlock()

	clientList := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clientList = append(clientList, client)
	}

	return clientList
}
