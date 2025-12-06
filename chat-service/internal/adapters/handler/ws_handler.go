package handler

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	ws "github.com/gorilla/websocket"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Configure allowed origins properly in production
		return true
	},
}

type WSHandler struct {
	manager     *websocket.ClientManager
	jwtSecret   string
	redisClient *redis.Client
}

func NewWSHandler(manager *websocket.ClientManager, jwtSecret string, client *redis.Client) *WSHandler {
	return &WSHandler{
		manager:     manager,
		jwtSecret:   jwtSecret,
		redisClient: client,
	}
}

// HandleConnection upgrades the HTTP connection to a WebSocket with JWT authentication
func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Extract token from query parameter (for WebSocket) or header
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		authHeader := r.Header.Get("Authorization")
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if tokenString == "" {
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		return
	}

	// Validate JWT token
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	exists, err := h.redisClient.Exists(r.Context(), "revoked:"+tokenString).Result()
	if err != nil {
		http.Error(w, "Redis error", http.StatusInternalServerError)
		return
	}

	if exists > 0 {
		http.Error(w, "Token revoked", http.StatusUnauthorized)
		return
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	// Get user_id from claims
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		http.Error(w, "Invalid user_id in token", http.StatusUnauthorized)
		return
	}
	userID := int64(userIDFloat)

	// Upgrade the connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	// Register the authenticated client
	h.manager.AddClient(userID, conn)
	log.Printf("User %d connected via WebSocket", userID)

	// Keep connection alive and listen for messages
	defer h.manager.RemoveClient(userID)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("User %d disconnected: %v", userID, err)
			break
		}
		// Handle incoming messages if needed
	}
}
