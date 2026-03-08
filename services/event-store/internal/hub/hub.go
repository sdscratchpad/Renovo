// Package hub manages active WebSocket connections and broadcasts change events
// to all connected clients whenever the event-store writes new data.
package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Event is the JSON payload pushed to every connected client.
type Event struct {
	Type string `json:"type"`
	At   string `json:"at"`
}

// Hub holds the set of live WebSocket connections and fans-out broadcasts.
type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

// New returns an initialised, empty Hub.
func New() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]struct{})}
}

// Register adds conn to the broadcast set.
func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

// Unregister removes conn from the broadcast set and closes it.
func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	conn.Close()
}

// Broadcast sends a JSON event to every registered client.
// Stale connections that fail to write are removed silently; the client is
// expected to reconnect automatically.
func (h *Hub) Broadcast(eventType string) {
	msg, _ := json.Marshal(Event{
		Type: eventType,
		At:   time.Now().UTC().Format(time.RFC3339Nano),
	})
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("hub: write to client: %v", err)
			delete(h.clients, conn)
			conn.Close()
		}
	}
}
