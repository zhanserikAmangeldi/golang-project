// package main

// import (
// 	"encoding/json"
// 	"log"
// 	"net/http"

// 	"chat-service/internal/adapters/handler"
// 	"chat-service/internal/adapters/repository"
// 	"chat-service/internal/adapters/websocket"
// 	"chat-service/chat-service/internal/core/service"

// 	"github.com/jmoiron/sqlx"
// 	_ "github.com/lib/pq"

// 	userpb  "github.com/zhanserikAmangeldi/proto/userpb"
// )

// func main() {
// 	httpPort := getEnv("HTTP_PORT", "8080")
// 	dbURL := getEnv("DB_URL", "postgres://chatuser:chatpass123@localhost:5432/chatdb?sslmode=disable")
// 	userServiceAddr := getEnv("USER_SERVICE_GRPC", "localhost:9091")

// 	// подключение к PostgreSQL (chatdb)
// 	db, err := sqlx.Connect("postgres", dbURL)
// 	if err != nil {
// 		log.Fatalln("failed to connect db:", err)
// 	}
// 	defer db.Close()
// 	log.Println("Connected to Chat Database")

// 	// gRPC client к user-service
// 	userConn, err := grpc.Dial(userServiceAddr, grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalln("failed to connect user-service:", err)
// 	}
// 	defer userConn.Close()
// 	userClient := userpb.NewUserServiceClient(userConn)

// 	wsManager := websocket.NewClientManager()
// 	repo := repository.NewPostgresRepository(db)

// 	chatService := service.NewChatService(repo, wsManager, userClient)

// 	wsHandler := handler.NewWSHandler(wsManager)
// 	http.HandleFunc("/ws", wsHandler.HandleConnection)

// 	type SendMessageRequest struct {
// 		SenderID    int64  `json:"sender_id"`
// 		RecipientID int64  `json:"recipient_id"`
// 		Content     string `json:"content"`
// 	}

// 	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		var req SendMessageRequest
// 		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 			http.Error(w, "Invalid JSON", http.StatusBadRequest)
// 			return
// 		}

// 		msg, err := chatService.SendMessage(context.Background(), req.SenderID, req.RecipientID, req.Content)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(msg)
// 	})

// 	addr := ":" + httpPort
// 	log.Println("Chat service started on", addr)
// 	if err := http.ListenAndServe(addr, nil); err != nil {
// 		log.Fatalln("Server failed:", err)
// 	}
// }

// func getEnv(key, def string) string {
// 	if v := os.Getenv(key); v != "" {
// 		return v
// 	}
// 	return def
// }

package main

import (
	"context"          // ← ДОБАВИТЬ
	"encoding/json"
	"log"
	"net/http"
	"os"               // ← ДОБАВИТЬ
	
	"chat-service/internal/adapters/handler"
	"chat-service/internal/adapters/repository"
	"chat-service/internal/adapters/websocket"
	"chat-service/internal/core/service"
	
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"                      // ← ДОБАВИТЬ
	"google.golang.org/grpc/credentials/insecure" // ← ДОБАВИТЬ
	
	userpb "github.com/zhanserikAmangeldi/proto/userpb"
)

func main() {
	httpPort := getEnv("HTTP_PORT", "8080")
	dbURL := getEnv("DB_URL", "postgres://chatuser:chatpass123@localhost:5432/chatdb?sslmode=disable")
	userServiceAddr := getEnv("USER_SERVICE_GRPC", "localhost:9091")

	// подключение к PostgreSQL (chatdb)
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalln("failed to connect db:", err)
	}
	defer db.Close()
	log.Println("Connected to Chat Database")

	// gRPC client к user-service
	userConn, err := grpc.Dial(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln("failed to connect user-service:", err)
	}
	defer userConn.Close()

	userClient := userpb.NewUserServiceClient(userConn)

	wsManager := websocket.NewClientManager()
	repo := repository.NewPostgresRepository(db)
	chatService := service.NewChatService(repo, wsManager, userClient)

	wsHandler := handler.NewWSHandler(wsManager)
	http.HandleFunc("/ws", wsHandler.HandleConnection)

	type SendMessageRequest struct {
		SenderID    int64  `json:"sender_id"`
		RecipientID int64  `json:"recipient_id"`
		Content     string `json:"content"`
	}

	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		msg, err := chatService.SendMessage(context.Background(), req.SenderID, req.RecipientID, req.Content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msg)
	})

	addr := ":" + httpPort
	log.Println("Chat service started on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalln("Server failed:", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}