package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	// Swagger - ВАЖНО: убедись, что путь к docs верный
	_ "github.com/zhanserikAmangeldi/chat-service/docs"
	httpSwagger "github.com/swaggo/http-swagger"

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

// @title Chat Service API
// @version 1.0
// @description API для работы с чатами, сообщениями и файлами (MinIO).
// @host localhost:8082
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	cfg := config.Load()

	// 1. Подключение к БД
	db, err := sqlx.Connect("postgres", cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// 2. Миграции
	log.Println("Running migrations...")
	if err := migration.AutoMigrate(cfg.GetDBURL()); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations applied successfully")

	// 3. Redis
	redisAddr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	redisClient := redisAdapter.NewRedisClient(redisAddr)
	log.Println("Connected to Redis")

	// 4. gRPC User Service
	userClient, err := grpcAdapter.NewUserClient(cfg.UserServiceURL)
	if err != nil {
		log.Fatalf("Failed to connect to User Service gRPC: %v", err)
	}
	defer userClient.Close()
	log.Println("Connected to User Service (gRPC)")

	// 5. MinIO
	minioService, err := service.NewMinioService(service.MinioConfig{
		Endpoint:  cfg.MinioHost + ":" + cfg.MinioApiPort,
		AccessKey: cfg.MinioAccessKey,
		SecretKey: cfg.MinioSecretKey,
		UseSSL:    cfg.MinioUseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MinIO: %v", err)
	}
	log.Println("Connected to MinIO")

	// 6. Инициализация слоев
	wsManager := websocket.NewClientManager()
	repo := repository.NewPostgresRepository(db)
	chatService := service.NewChatService(repo, redisClient, userClient)

	go background.StartRedisListener(context.Background(), redisClient, wsManager)

	wsHandler := handler.NewWSHandler(wsManager, cfg.JWTSecret)
	chatHandler := handler.NewChatHandler(chatService)
	fileHandler := handler.NewFileHandler(minioService, chatService)

	// 7. Настройка РОУТЕРА (Mux)
	mux := http.NewServeMux()

	// --- РЕГИСТРАЦИЯ ПУТЕЙ В MUX ---

	// Swagger (Исправлено: регистрируем префикс без лишних глобальных вызовов)
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%s/swagger/doc.json", cfg.HTTPPort)),
	))

	// WebSocket
	mux.HandleFunc("/ws", wsHandler.HandleConnection)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "healthy",
			"service":  "chat-service",
			"database": "connected",
		})
	})

	// Middlewares
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	// API Эндпоинты
	mux.Handle("/api/v1/groups/create", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID, _ := middleware.GetUserID(r)
		var req struct {
			Name      string  `json:"name"`
			MemberIDs []int64 `json:"member_ids"`
		}
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
	})))

	mux.Handle("/api/v1/messages/send", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID, _ := middleware.GetUserID(r)
		var req struct {
			RecipientID    int64   `json:"recipient_id"`
			ConversationID int64   `json:"conversation_id"`
			Content        string  `json:"content"`
			MessageType    string  `json:"message_type"`
			FileURL        *string `json:"file_url,omitempty"`
			FileName       *string `json:"file_name,omitempty"`
			MimeType       *string `json:"mime_type,omitempty"`
			FileSize       *int64  `json:"file_size,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		msg, err := chatService.SendMessage(r.Context(), userID, req.RecipientID, req.Content, req.ConversationID, req.MessageType, req.FileURL, req.FileName, req.MimeType, req.FileSize)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msg)
	})))

	// Остальные роуты
	mux.Handle("/api/v1/messages/history", authMiddleware(http.HandlerFunc(chatHandler.GetHistory)))
	mux.Handle("/api/v1/conversations", authMiddleware(http.HandlerFunc(chatHandler.GetConversations)))
	mux.Handle("/api/v1/messages/read", authMiddleware(http.HandlerFunc(chatHandler.MarkAsRead)))
	mux.Handle("/api/v1/messages/reactions/add", authMiddleware(http.HandlerFunc(chatHandler.AddReaction)))
	mux.Handle("/api/v1/messages/reactions/remove", authMiddleware(http.HandlerFunc(chatHandler.RemoveReaction)))
	mux.Handle("/api/v1/messages/edit", authMiddleware(http.HandlerFunc(chatHandler.EditMessage)))
	mux.Handle("/api/v1/messages/delete", authMiddleware(http.HandlerFunc(chatHandler.DeleteMessage)))

	mux.Handle("/api/v1/files/upload", authMiddleware(http.HandlerFunc(fileHandler.UploadFile)))
	mux.Handle("/api/v1/files/send", authMiddleware(http.HandlerFunc(fileHandler.SendMessageWithFile)))
	mux.Handle("/api/v1/files/get", authMiddleware(http.HandlerFunc(fileHandler.GetFile)))

	// 8. ЗАПУСК
	serverAddr := fmt.Sprintf(":%s", cfg.HTTPPort)
	log.Printf("Chat service starting on port %s", cfg.HTTPPort)
	
	// ВАЖНО: используем mux, а не nil!
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalln("Server failed:", err)
	}
}