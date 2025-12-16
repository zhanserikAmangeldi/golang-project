package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/zhanserikAmangeldi/chat-service/config"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/background"
	grpcAdapter "github.com/zhanserikAmangeldi/chat-service/internal/adapters/grpc"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/handler"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/repository"
	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/websocket"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/service"
	"github.com/zhanserikAmangeldi/chat-service/internal/middleware"
	"github.com/zhanserikAmangeldi/chat-service/internal/migration"
	redisAdapter "github.com/zhanserikAmangeldi/chat-service/internal/redis"
)

func main() {
	cfg := config.Load()

	// 1. Connect to PostgreSQL
	db, err := sqlx.Connect("postgres", cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// 2. Run Migrations
	log.Println("Running migrations...")
	if err := migration.AutoMigrate(cfg.GetDBURL()); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations applied successfully")

	redisAddr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)

	redisClient := redisAdapter.NewRedisClient(redisAddr)
	log.Println("Connected to Redis")

	userClient, err := grpcAdapter.NewUserClient(cfg.UserServiceURL)
	if err != nil {
		log.Fatalf("Failed to connect to User Service gRPC: %v", err)
	}
	defer userClient.Close()
	log.Println("Connected to User Service (gRPC)")

	wsManager := websocket.NewClientManager()
	repo := repository.NewPostgresRepository(db)

	chatService := service.NewChatService(repo, redisClient, userClient)

	go background.StartRedisListener(context.Background(), redisClient, wsManager)

	wsHandler := handler.NewWSHandler(wsManager, cfg.JWTSecret)
	http.HandleFunc("/ws", wsHandler.HandleConnection)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "healthy",
			"service":  "chat-service",
			"database": "connected",
			"redis":    "connected",
			"grpc":     "connected",
		})
	})

	mux := http.NewServeMux()

	createGroupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}

		type CreateGroupRequest struct {
			Name      string  `json:"name"`
			MemberIDs []int64 `json:"member_ids"`
		}

		var req CreateGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		conv, err := chatService.CreateGroup(r.Context(), req.Name, userID, req.MemberIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(conv)
	})

	sendMessageHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}

		type SendMessageRequest struct {
			RecipientID    int64  `json:"recipient_id"`
			ConversationID int64  `json:"conversation_id"`
			Content        string `json:"content"`
		}

		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		msg, err := chatService.SendMessage(r.Context(), userID, req.RecipientID, req.Content, req.ConversationID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msg)
	})

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

		_ = userID

		messages, err := chatService.GetHistory(r.Context(), conversationID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	})

	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	mux.Handle("/api/v1/groups/create", authMiddleware(createGroupHandler))
	mux.Handle("/api/v1/messages/send", authMiddleware(sendMessageHandler))
	mux.Handle("/api/v1/messages/history", authMiddleware(getHistoryHandler))

	http.Handle("/api/", mux)

	log.Printf("Chat service starting on port %s", cfg.HTTPPort)
	log.Println("Features: Redis Pub/Sub [ON], gRPC User Validation [ON]")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.HTTPPort), nil); err != nil {
		log.Fatalln("Server failed:", err)
	}
}
