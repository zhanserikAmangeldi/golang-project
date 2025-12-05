package ports

import (
	"context"

	"chat-service/internal/core/domain"
)

type ChatRepository interface {
	// Conversation methods
	CreateConversation(ctx context.Context, conv *domain.Conversation) error
	GetConversationByID(ctx context.Context, id int64) (*domain.Conversation, error)

	// This is crucial for 1:1 chat: Find if a chat already exists between two users
	FindOneToOneConversation(ctx context.Context, user1, user2 int64) (*domain.Conversation, error)

	// Participant methods
	AddParticipant(ctx context.Context, part *domain.Participant) error

	// Message methods
	SaveMessage(ctx context.Context, msg *domain.Message) error
	GetMessages(ctx context.Context, conversationID int64, limit, offset int) ([]domain.Message, error)
}
