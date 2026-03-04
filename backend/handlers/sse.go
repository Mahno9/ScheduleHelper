package handlers

import (
	"fmt"
	"net/http"
	"sync"
)

var clients = make(map[chan string]bool)
var clientsMutex sync.Mutex

func Broadcast(event string, data string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)
	for client := range clients {
		client <- msg
	}
}

func SSEHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	clientChan := make(chan string)
	clientsMutex.Lock()
	clients[clientChan] = true
	clientsMutex.Unlock()

	defer func() {
		clientsMutex.Lock()
		delete(clients, clientChan)
		clientsMutex.Unlock()
		close(clientChan)
	}()

	for {
		select {
		case msg := <-clientChan:
			fmt.Fprint(w, msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}