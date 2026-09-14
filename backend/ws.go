package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Event is broadcast to every connected frontend client over WebSocket.
// Type is one of "sync" or "playlist_updated".
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// SyncPayload is sent with a "sync" event.
type SyncPayload struct {
	MediaItem       MediaItem `json:"mediaItem"`
	DurationSeconds int       `json:"durationSeconds"`
}

// PlaylistUpdatedPayload is sent with a "playlist_updated" event.
type PlaylistUpdatedPayload struct {
	Window Window `json:"window"`
}

// Hub keeps track of connected WebSocket clients and broadcasts events to
// all of them. This is how every window (even across different browser
// tabs) finds out about a sync trigger or a playlist change at the same
// moment, without needing to poll the backend.
type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]bool)}
}

var upgrader = websocket.Upgrader{
	// Allow connections from any origin: the frontend and backend are
	// deployed separately (e.g. Vercel + Render), so a strict same-origin
	// check would block the real frontend too.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandleWS upgrades an HTTP connection to a WebSocket and registers it
// with the hub until the client disconnects.
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	// We don't expect messages from the client on this connection, but we
	// still need to read (and discard) frames so the connection's ping/pong
	// and close handling work correctly, and so we notice disconnects.
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

// Broadcast sends an event to every currently connected client.
func (h *Hub) Broadcast(event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteJSON(event); err != nil {
			conn.Close()
			delete(h.clients, conn)
		}
	}
}
