package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	ws "github.com/gorilla/websocket"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/websocket"
)

type WSHandler struct {
	manager   *websocket.ClientManager
	jwtSecret string
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWSHandler(manager *websocket.ClientManager, jwtSecret string) *WSHandler {
	return &WSHandler{
		manager:   manager,
		jwtSecret: jwtSecret,
	}
}

func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	var userID int64
	tokenString := r.URL.Query().Get("token")

	idString := r.URL.Query().Get("id")

	if tokenString != "" {
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Token missing user_id", http.StatusUnauthorized)
			return
		}
		userID = int64(userIDFloat)

	} else if idString != "" {
		var err error
		userID, err = strconv.ParseInt(idString, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}
		log.Printf("[WS] ⚠️ WARNING: Connecting via insecure ID parameter: %d", userID)

	} else {
		http.Error(w, "Missing token or id", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[WS] Failed to upgrade:", err)
		return
	}

	h.manager.AddClient(userID, conn)
	log.Printf("[WS] User connected: %d", userID)

	defer h.manager.RemoveClient(userID)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WS] User %d disconnected", userID)
			break
		}
	}
}
