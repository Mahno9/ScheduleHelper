package services

import (
	"encoding/json"
	"log"
	"sync"
)

// SSEHub manages Server-Sent Events connections
type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

// NewSSEHub creates a new hub
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan []byte]struct{}),
	}
}

// Subscribe returns a channel that will receive SSE messages
func (h *SSEHub) Subscribe() chan []byte {
	ch := make(chan []byte, 64)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes a client channel
func (h *SSEHub) Unsubscribe(ch chan []byte) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	// drain to avoid goroutine leaks
	for len(ch) > 0 {
		<-ch
	}
	close(ch)
}

// Broadcast sends a typed event to all connected clients
func (h *SSEHub) Broadcast(eventType string, payload interface{}) {
	msg := map[string]interface{}{
		"type":    eventType,
		"payload": payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("SSEHub marshal error: %v", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- data:
		default:
			log.Printf("SSEHub: client slow, dropping message")
		}
	}
}
