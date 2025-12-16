package ports

import (
	"context"

	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
)

type ChatRepository interface {
	CreateConversation(ctx context.Context, conv *model.Conversation) error
	GetConversationByID(ctx context.Context, id int64) (*model.Conversation, error)
	FindOneToOneConversation(ctx context.Context, user1, user2 int64) (*model.Conversation, error)
	AddParticipant(ctx context.Context, part *model.Participant) error
	SaveMessage(ctx context.Context, msg *model.Message) error
	GetMessages(ctx context.Context, conversationID int64, limit, offset int) ([]model.Message, error)
	GetParticipants(ctx context.Context, conversationID int64) ([]int64, error)
	IsParticipant(ctx context.Context, user1, user2 int64) (bool, error)
}
