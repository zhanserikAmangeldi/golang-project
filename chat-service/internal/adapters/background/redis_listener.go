package background

import (
	"context"
	"encoding/json"
	"log"

	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/websocket"
	"github.com/zhanserikAmangeldi/chat-service/internal/redis"
)

// StartRedisListener runs in the background.
// It subscribes to Redis and checks if any recipients are connected to THIS instance.
func StartRedisListener(ctx context.Context, redisClient *redis.RedisClient, wsManager *websocket.ClientManager) {
	log.Println("Started Redis Subscriber...")

	msgChan := redisClient.Subscribe(ctx)

	for payload := range msgChan {
		msgBytes, _ := json.Marshal(payload.Message)

		for _, userID := range payload.RecipientIDs {
			if conn, ok := wsManager.GetClient(userID); ok {
				if err := conn.WriteMessage(1, msgBytes); err != nil {
					log.Printf("Failed to write to WS for user %d", userID)
				}
			}
		}
	}
}
