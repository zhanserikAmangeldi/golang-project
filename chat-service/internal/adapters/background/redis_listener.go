package background

import (
	"context"
	"encoding/json"
	"log"

	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/websocket"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
	"github.com/zhanserikAmangeldi/chat-service/internal/redis"
)

func StartRedisListener(ctx context.Context, redisClient *redis.RedisClient, wsManager *websocket.ClientManager) {
	log.Println("Started Redis Subscriber...")

	msgChan := redisClient.Subscribe(ctx)

	for payload := range msgChan {
		handleBroadcastMessage(payload, wsManager)
	}
}

func handleBroadcastMessage(payload redis.BroadcastMessage, wsManager *websocket.ClientManager) {
	switch payload.Type {
	case "message":
		if payload.Message != nil {
			sendToRecipients(wsManager, payload.RecipientIDs, model.WSMessage{
				Type:    "message",
				Payload: payload.Message,
			})
		}

	case "message_edit":
		if payload.Message != nil {
			sendToRecipients(wsManager, payload.RecipientIDs, model.WSMessage{
				Type:    "message_edit",
				Payload: payload.Message,
			})
		}

	case "message_delete":
		sendToRecipients(wsManager, payload.RecipientIDs, model.WSMessage{
			Type:    "message_delete",
			Payload: payload.Payload,
		})

	case "typing":
		sendToRecipients(wsManager, payload.RecipientIDs, model.WSMessage{
			Type:    "typing",
			Payload: payload.Payload,
		})

	case "status":
		sendToRecipients(wsManager, payload.RecipientIDs, model.WSMessage{
			Type:    "status",
			Payload: payload.Payload,
		})

	case "reaction_add":
		sendToRecipients(wsManager, payload.RecipientIDs, model.WSMessage{
			Type:    "reaction_add",
			Payload: payload.Payload,
		})

	case "reaction_remove":
		sendToRecipients(wsManager, payload.RecipientIDs, model.WSMessage{
			Type:    "reaction_remove",
			Payload: payload.Payload,
		})

	case "read_receipt":
		sendToRecipients(wsManager, payload.RecipientIDs, model.WSMessage{
			Type:    "read_receipt",
			Payload: payload.Payload,
		})

	default:
		log.Printf("Unknown broadcast type: %s", payload.Type)
	}
}

func sendToRecipients(wsManager *websocket.ClientManager, recipientIDs []int64, message model.WSMessage) {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	for _, userID := range recipientIDs {
		if conn, ok := wsManager.GetClient(userID); ok {
			if err := conn.WriteMessage(1, msgBytes); err != nil {
				log.Printf("Failed to write to WS for user %d: %v", userID, err)
			}
		}
	}
}
