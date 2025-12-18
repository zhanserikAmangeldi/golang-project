package redis

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
)

type IRedisClient interface {
	Publish(ctx context.Context, msg model.Message, recipients []int64) error
	PublishTyping(ctx context.Context, event model.TypingEvent, recipients []int64) error
	PublishStatus(ctx context.Context, event model.OnlineStatusEvent, recipients []int64) error
	PublishReaction(ctx context.Context, reaction model.Reaction, recipients []int64) error
	PublishReactionRemoval(ctx context.Context, reaction model.Reaction, recipients []int64) error
	PublishReadReceipt(ctx context.Context, readReceipt model.MessageRead, recipients []int64) error
	PublishMessageEdit(ctx context.Context, msg model.Message, recipients []int64) error
	PublishMessageDeletion(ctx context.Context, messageID int64, recipients []int64) error
	Subscribe(ctx context.Context) <-chan BroadcastMessage
}

const (
	ChannelMessage     = "chat.message"
	ChannelTyping      = "chat.typing"
	ChannelStatus      = "chat.status"
	ChannelReaction    = "chat.reaction"
	ChannelReadReceipt = "chat.read_receipt"
)

type BroadcastMessage struct {
	Type         string         `json:"type"` // message, typing, status, reaction, read_receipt, message_edit, message_delete
	Message      *model.Message `json:"message,omitempty"`
	RecipientIDs []int64        `json:"recipient_ids"`
	Payload      interface{}    `json:"payload,omitempty"`
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

func (r *RedisClient) Publish(ctx context.Context, msg model.Message, recipients []int64) error {
	payload := BroadcastMessage{
		Type:         "message",
		Message:      &msg,
		RecipientIDs: recipients,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.client.Publish(ctx, ChannelMessage, data).Err()
}

func (r *RedisClient) PublishTyping(ctx context.Context, event model.TypingEvent, recipients []int64) error {
	payload := BroadcastMessage{
		Type:         "typing",
		RecipientIDs: recipients,
		Payload:      event,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.client.Publish(ctx, ChannelTyping, data).Err()
}

func (r *RedisClient) PublishStatus(ctx context.Context, event model.OnlineStatusEvent, recipients []int64) error {
	payload := BroadcastMessage{
		Type:         "status",
		RecipientIDs: recipients,
		Payload:      event,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.client.Publish(ctx, ChannelStatus, data).Err()
}

func (r *RedisClient) PublishReaction(ctx context.Context, reaction model.Reaction, recipients []int64) error {
	payload := BroadcastMessage{
		Type:         "reaction_add",
		RecipientIDs: recipients,
		Payload:      reaction,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.client.Publish(ctx, ChannelReaction, data).Err()
}

func (r *RedisClient) PublishReactionRemoval(ctx context.Context, reaction model.Reaction, recipients []int64) error {
	payload := BroadcastMessage{
		Type:         "reaction_remove",
		RecipientIDs: recipients,
		Payload:      reaction,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.client.Publish(ctx, ChannelReaction, data).Err()
}

func (r *RedisClient) PublishReadReceipt(ctx context.Context, readReceipt model.MessageRead, recipients []int64) error {
	payload := BroadcastMessage{
		Type:         "read_receipt",
		RecipientIDs: recipients,
		Payload:      readReceipt,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.client.Publish(ctx, ChannelReadReceipt, data).Err()
}

func (r *RedisClient) PublishMessageEdit(ctx context.Context, msg model.Message, recipients []int64) error {
	payload := BroadcastMessage{
		Type:         "message_edit",
		Message:      &msg,
		RecipientIDs: recipients,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.client.Publish(ctx, ChannelMessage, data).Err()
}

func (r *RedisClient) PublishMessageDeletion(ctx context.Context, messageID int64, recipients []int64) error {
	payload := BroadcastMessage{
		Type:         "message_delete",
		RecipientIDs: recipients,
		Payload:      map[string]int64{"message_id": messageID},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.client.Publish(ctx, ChannelMessage, data).Err()
}

func (r *RedisClient) Subscribe(ctx context.Context) <-chan BroadcastMessage {
	ch := make(chan BroadcastMessage)

	pubsub := r.client.Subscribe(ctx, ChannelMessage, ChannelTyping, ChannelStatus, ChannelReaction, ChannelReadReceipt)

	go func() {
		defer close(ch)
		defer pubsub.Close()

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
