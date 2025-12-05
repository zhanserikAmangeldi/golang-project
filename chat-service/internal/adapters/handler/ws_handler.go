package handler

import (
	"log"
	"net/http"
	"strconv"

	ws "github.com/gorilla/websocket"
	"chat-service/internal/adapters/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for now (since we have no frontend)
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	manager *websocket.ClientManager
}

func NewWSHandler(manager *websocket.ClientManager) *WSHandler {
	return &WSHandler{manager: manager}
}

// HandleConnection upgrades the HTTP connection to a WebSocket
func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// 1. Extract User ID from URL (Temporary fake auth)
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	// 2. Upgrade the connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	// 3. Register the client
	h.manager.AddClient(userID, conn)
	log.Printf("User connected: %s", userID)

	// 4. Listen for incoming messages (Keep the connection alive)
	// In a real app, we might read messages here. For now, we just keep it open.
	// If the loop breaks, user disconnected.
	defer h.manager.RemoveClient(userID)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("User disconnected: %s", userID)
			break
		}
	}
}
