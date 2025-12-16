package redis

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
)

const ChannelName = "chat.broadcast"

// BroadcastMessage is what we send over Redis
type BroadcastMessage struct {
	Message      model.Message `json:"message"`
	RecipientIDs []int64       `json:"recipient_ids"` // List of users who should receive this
}

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

// Publish sends the message to the Redis Channel
func (r *RedisClient) Publish(ctx context.Context, msg model.Message, recipients []int64) error {
	payload := BroadcastMessage{
		Message:      msg,
		RecipientIDs: recipients,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.client.Publish(ctx, ChannelName, data).Err()
}

// Subscribe returns a Go channel that receives messages from Redis
func (r *RedisClient) Subscribe(ctx context.Context) <-chan BroadcastMessage {
	// Create a channel to pass data back to the caller
	ch := make(chan BroadcastMessage)

	pubsub := r.client.Subscribe(ctx, ChannelName)

	// Run a goroutine to listen to Redis
	go func() {
		defer close(ch)
		defer pubsub.Close()

		// Read the channel
		chRed := pubsub.Channel()

		for msg := range chRed {
			var broadcastMsg BroadcastMessage
			if err := json.Unmarshal([]byte(msg.Payload), &broadcastMsg); err != nil {
				log.Println("Error unmarshaling redis msg:", err)
				continue
			}
			ch <- broadcastMsg
		}
	}()

	return ch
}
