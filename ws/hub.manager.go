package ws

import "sync"

type HubManager struct {
	mu   sync.RWMutex
	hubs map[string]*Hub
}

var manager = &HubManager{
	hubs: make(map[string]*Hub),
}

// GetHub returns an existing hub by name or creates a new one if not present.
func GetHub(name string) *Hub {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if h, ok := manager.hubs[name]; ok {
		return h
	}
	h := NewHub()
	manager.hubs[name] = h
	go h.run()
	return h
}

// Broadcast sends a message to the named hub (if it exists)
func Broadcast(name string, msg []byte) {
	h := GetHub(name) // ensures hub exists
	h.broadcast <- msg
}
