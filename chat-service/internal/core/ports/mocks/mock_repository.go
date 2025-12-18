package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
)

type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) CreateConversation(ctx context.Context, conv *model.Conversation) error {
	args := m.Called(ctx, conv)
	conv.ID = 1
	return args.Error(0)
}

func (m *MockChatRepository) GetConversationByID(ctx context.Context, id int64) (*model.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Conversation), args.Error(1)
}

func (m *MockChatRepository) FindOneToOneConversation(ctx context.Context, user1, user2 int64) (*model.Conversation, error) {
	args := m.Called(ctx, user1, user2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Conversation), args.Error(1)
}

func (m *MockChatRepository) GetUserConversations(ctx context.Context, userID int64, limit, offset int) ([]model.ConversationWithLastMessage, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]model.ConversationWithLastMessage), args.Error(1)
}

func (m *MockChatRepository) AddParticipant(ctx context.Context, part *model.Participant) error {
	args := m.Called(ctx, part)
	return args.Error(0)
}

func (m *MockChatRepository) GetParticipants(ctx context.Context, conversationID int64) ([]int64, error) {
	args := m.Called(ctx, conversationID)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockChatRepository) IsParticipant(ctx context.Context, convID, userID int64) (bool, error) {
	args := m.Called(ctx, convID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockChatRepository) SaveMessage(ctx context.Context, msg *model.Message) error {
	args := m.Called(ctx, msg)
	msg.ID = 42
	return args.Error(0)
}

func (m *MockChatRepository) GetMessages(ctx context.Context, conversationID int64, limit, offset int) ([]model.Message, error) {
	args := m.Called(ctx, conversationID, limit, offset)
	return args.Get(0).([]model.Message), args.Error(1)
}

func (m *MockChatRepository) GetMessageByID(ctx context.Context, messageID int64) (*model.Message, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Message), args.Error(1)
}

func (m *MockChatRepository) GetLastMessage(ctx context.Context, conversationID int64) (*model.Message, error) {
	args := m.Called(ctx, conversationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Message), args.Error(1)
}

func (m *MockChatRepository) EditMessage(ctx context.Context, messageID int64, newContent string) error {
	args := m.Called(ctx, messageID, newContent)
	return args.Error(0)
}

func (m *MockChatRepository) DeleteMessage(ctx context.Context, messageID int64) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *MockChatRepository) MarkMessageAsRead(ctx context.Context, messageID, userID int64) error {
	args := m.Called(ctx, messageID, userID)
	return args.Error(0)
}

func (m *MockChatRepository) GetMessageReads(ctx context.Context, messageID int64) ([]int64, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockChatRepository) AddReaction(ctx context.Context, messageID, userID int64, reaction string) error {
	args := m.Called(ctx, messageID, userID, reaction)
	return args.Error(0)
}

func (m *MockChatRepository) RemoveReaction(ctx context.Context, messageID, userID int64, reaction string) error {
	args := m.Called(ctx, messageID, userID, reaction)
	return args.Error(0)
}

func (m *MockChatRepository) GetMessageReactions(ctx context.Context, messageID int64) ([]model.Reaction, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).([]model.Reaction), args.Error(1)
}
