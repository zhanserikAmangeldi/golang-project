package ports

import (
	"context"

	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
)

type ChatRepository interface {
	// Conversations
	CreateConversation(ctx context.Context, conv *model.Conversation) error
	GetConversationByID(ctx context.Context, id int64) (*model.Conversation, error)
	FindOneToOneConversation(ctx context.Context, user1, user2 int64) (*model.Conversation, error)
	GetUserConversations(ctx context.Context, userID int64, limit, offset int) ([]model.ConversationWithLastMessage, error)

	// Participants
	AddParticipant(ctx context.Context, part *model.Participant) error
	GetParticipants(ctx context.Context, conversationID int64) ([]int64, error)
	IsParticipant(ctx context.Context, convID, userID int64) (bool, error)

	// Messages
	SaveMessage(ctx context.Context, msg *model.Message) error
	GetMessages(ctx context.Context, conversationID int64, limit, offset int) ([]model.Message, error)
	GetMessageByID(ctx context.Context, messageID int64) (*model.Message, error)
	GetLastMessage(ctx context.Context, conversationID int64) (*model.Message, error)
	EditMessage(ctx context.Context, messageID int64, newContent string) error
	DeleteMessage(ctx context.Context, messageID int64) error

	// Read Receipts
	MarkMessageAsRead(ctx context.Context, messageID, userID int64) error
	GetMessageReads(ctx context.Context, messageID int64) ([]int64, error)

	// Reactions
	AddReaction(ctx context.Context, messageID, userID int64, reaction string) error
	RemoveReaction(ctx context.Context, messageID, userID int64, reaction string) error
	GetMessageReactions(ctx context.Context, messageID int64) ([]model.Reaction, error)
}
