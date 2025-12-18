package handler

import (
	"context"
	"encoding/json"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/ports"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	ws "github.com/gorilla/websocket"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/websocket"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
	redisAdapter "github.com/zhanserikAmangeldi/chat-service/internal/redis"
)

type WSHandler struct {
	manager     *websocket.ClientManager
	jwtSecret   string
	redisClient *redisAdapter.RedisClient
	chatRepo    ports.ChatRepository
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWSHandler(manager *websocket.ClientManager, jwtSecret string, redisClient *redisAdapter.RedisClient, chatRepo ports.ChatRepository) *WSHandler {
	return &WSHandler{
		manager:     manager,
		jwtSecret:   jwtSecret,
		redisClient: redisClient,
		chatRepo:    chatRepo,
	}
}

func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	var userID int64
	tokenString := r.URL.Query().Get("token")

	if tokenString == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

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

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[WS] Failed to upgrade:", err)
		return
	}

	h.manager.AddClient(userID, conn)
	log.Printf("[WS] User connected: %d", userID)

	defer func() {
		h.manager.RemoveClient(userID)
		log.Printf("[WS] User disconnected: %d", userID)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WS] User %d disconnected: %v", userID, err)
			break
		}

		var wsMsg model.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("[WS] Invalid message format from user %d: %v", userID, err)
			continue
		}

		h.handleWSMessage(r.Context(), userID, wsMsg)
	}
}

func (h *WSHandler) handleWSMessage(ctx context.Context, userID int64, msg model.WSMessage) {
	switch msg.Type {
	case "typing":
		var typingEvent model.TypingEvent
		data, _ := json.Marshal(msg.Payload)
		if err := json.Unmarshal(data, &typingEvent); err != nil {
			log.Printf("[WS] Invalid typing event: %v", err)
			return
		}

		typingEvent.UserID = userID

		// TODO: Get conversation participants
		recipients, err := h.chatRepo.GetParticipants(ctx, typingEvent.ConversationID)
		if err != nil {
			log.Printf("[WS] Failed to get recipients: %v", err)
			return
		}

		_ = h.redisClient.PublishTyping(ctx, typingEvent, recipients)

	case "status":
		var statusEvent model.OnlineStatusEvent
		data, _ := json.Marshal(msg.Payload)
		if err := json.Unmarshal(data, &statusEvent); err != nil {
			log.Printf("[WS] Invalid status event: %v", err)
			return
		}

		statusEvent.UserID = userID

		// TODO: Broadcast to user's contacts/chat participants
		conversations, err := h.chatRepo.GetUserConversations(ctx, userID, 100, 0)
		if err != nil {
			log.Printf("[WS] Failed to get user conversations: %v", err)
			return
		}

		var recipients []int64
		for _, conv := range conversations {
			parts, _ := h.chatRepo.GetParticipants(ctx, conv.ID)
			for _, p := range parts {
				if p != userID {
					recipients = append(recipients, p)
				}
			}
		}

		_ = h.redisClient.PublishStatus(ctx, statusEvent, recipients)
	}
}
