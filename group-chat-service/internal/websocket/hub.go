package websocket

import (
	"sync"
)

// Message represents the payload sent over WS
type Message struct {
	GroupID  int    `json:"group_id"`
	SenderID int    `json:"sender_id"`
	Content  string `json:"content"`
}

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Rooms maps GroupID to a set of Clients (Active members in that group)
	rooms map[int]map[*Client]bool

	// Mutex to protect map access
	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		rooms:      make(map[int]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			// Add client to their subscribed group rooms
			if _, ok := h.rooms[client.GroupID]; !ok {
				h.rooms[client.GroupID] = make(map[*Client]bool)
			}
			h.rooms[client.GroupID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				// Remove from specific room
				if members, ok := h.rooms[client.GroupID]; ok {
					delete(members, client)
					if len(members) == 0 {
						delete(h.rooms, client.GroupID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			// Send only to clients in the specific group
			if members, ok := h.rooms[msg.GroupID]; ok {
				for client := range members {
					// Don't send back to sender if preferred, or send to all
					select {
					case client.send <- msg:
					default:
						close(client.send)
						delete(h.clients, client)
						delete(members, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}
