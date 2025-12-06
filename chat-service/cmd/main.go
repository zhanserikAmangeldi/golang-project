package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/zhanserikAmangeldi/chat-service/config"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/handler"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/repository"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/websocket"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/service"
	"github.com/zhanserikAmangeldi/chat-service/internal/middleware"
	"github.com/zhanserikAmangeldi/chat-service/internal/migration"
)

func main() {
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := sqlx.Connect("postgres", cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Run database migrations
	log.Println("Running migrations...")
	if err := migration.AutoMigrate(cfg.GetDBURL()); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations applied successfully")

	redisClient := NewRedisClient(cfg)
	defer redisClient.Close()

	// Initialize components
	wsManager := websocket.NewClientManager()
	repo := repository.NewPostgresRepository(db)
	chatService := service.NewChatService(repo, wsManager)

	// WebSocket handler with simple JWT authentication (no Redis)
	wsHandler := handler.NewWSHandler(wsManager, cfg.JWTSecret, redisClient)
	http.HandleFunc("/ws", wsHandler.HandleConnection)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "healthy",
			"service":  "chat-service",
			"database": "connected",
		})
	})

	// Protected endpoints
	mux := http.NewServeMux()

	// Send message endpoint
	sendMessageHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get authenticated user ID from context
		userID, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}

		type SendMessageRequest struct {
			RecipientID int64  `json:"recipient_id"`
			Content     string `json:"content"`
		}

		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		_, tokenString, err := middleware.GetToken(r, cfg.JWTSecret)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		exists, err := RecipientExists(r.Context(), req.RecipientID, cfg.UserServiceURL, tokenString)
		if err != nil {
			http.Error(w, "Failed to validate recipient", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Recipient does not exist", http.StatusBadRequest)
			return
		}

		// Use authenticated user as sender
		msg, err := chatService.SendMessage(r.Context(), userID, req.RecipientID, req.Content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msg)
	})

	// Get conversation history endpoint
	getHistoryHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}

		conversationIDStr := r.URL.Query().Get("conversation_id")
		if conversationIDStr == "" {
			http.Error(w, "conversation_id required", http.StatusBadRequest)
			return
		}

		var conversationID int64
		fmt.Sscanf(conversationIDStr, "%d", &conversationID)

		// TODO: Verify user is participant in this conversation
		ok, err = repo.IsParticipant(r.Context(), conversationID, userID)
		if err != nil {
			http.Error(w, "Failed to validate is participant conversation", http.StatusInternalServerError)
			return
		}

		if !ok {
			http.Error(w, "You do not have permission to view this conversation", http.StatusForbidden)
			return
		}

		messages, err := chatService.GetHistory(r.Context(), conversationID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	})

	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret, redisClient)
	mux.Handle("/api/v1/messages/send", authMiddleware(sendMessageHandler))
	mux.Handle("/api/v1/messages/history", authMiddleware(getHistoryHandler))

	http.Handle("/api/", mux)

	log.Printf("Chat service starting on port %s", cfg.HTTPPort)
	log.Println("WebSocket endpoint: /ws?token=<JWT_TOKEN>")
	log.Println("Send message: POST /api/v1/messages/send")
	log.Println("Get history: GET /api/v1/messages/history?conversation_id=<ID>")
	log.Println("")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.HTTPPort), nil); err != nil {
		log.Fatalln("Server failed:", err)
	}
}

func NewRedisClient(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: cfg.GetRedisAddr(),
		DB:   0,
	})
}

func RecipientExists(ctx context.Context, recipientID int64, userServiceURL string, token string) (bool, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/users/%d", userServiceURL, recipientID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, fmt.Errorf("unexpected status code %d", resp.StatusCode)
}
