package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/websocket"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
)

type MockRepo struct{ mock.Mock }

func (m *MockRepo) GetConversationByID(ctx context.Context, id int64) (*model.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Conversation), args.Error(1)
}

func (m *MockRepo) FindOneToOneConversation(ctx context.Context, s, r int64) (*model.Conversation, error) {
	args := m.Called(ctx, s, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Conversation), args.Error(1)
}

func (m *MockRepo) CreateConversation(ctx context.Context, c *model.Conversation) error {
	args := m.Called(ctx, c)
	// Simulate DB assigning ID
	if c.ID == 0 {
		c.ID = 777
	}
	return args.Error(0)
}

func (m *MockRepo) AddParticipant(ctx context.Context, p *model.Participant) error {
	return m.Called(ctx, p).Error(0)
}

func (m *MockRepo) SaveMessage(ctx context.Context, msg *model.Message) error {
	return m.Called(ctx, msg).Error(0)
}

func (m *MockRepo) GetMessages(ctx context.Context, id int64, limit, offset int) ([]model.Message, error) {
	args := m.Called(ctx, id, limit, offset)
	return args.Get(0).([]model.Message), args.Error(1)
}

func TestSendMessage_ExistingConversation(t *testing.T) {
	ctx := context.Background()

	repo := new(MockRepo)
	wsMgr := websocket.NewClientManager()
	svc := NewChatService(repo, wsMgr)

	sender := int64(1)
	recipient := int64(2)
	content := "Hello!"

	conv := &model.Conversation{ID: 10}

	// CORRECT: mock.Anything only for context, specific values for IDs
	repo.On("FindOneToOneConversation", mock.Anything, sender, recipient).
		Return(conv, nil)

	repo.On("SaveMessage", mock.Anything, mock.AnythingOfType("*domain.Message")).
		Return(nil)

	msg, err := svc.SendMessage(ctx, sender, recipient, content)

	assert.NoError(t, err)
	assert.Equal(t, conv.ID, msg.ConversationID)
	assert.Equal(t, content, msg.Content)

	repo.AssertExpectations(t)
}

func TestSendMessage_NoConversation_CreatesConversation(t *testing.T) {
	ctx := context.Background()

	repo := new(MockRepo)
	wsMgr := websocket.NewClientManager()
	svc := NewChatService(repo, wsMgr)

	sender := int64(1)
	recipient := int64(2)
	content := "Hi!"

	// CORRECT: Use mock.Anything only for context
	repo.On("FindOneToOneConversation", mock.Anything, sender, recipient).
		Return((*model.Conversation)(nil), nil)

	// CreateConversation: context + pointer
	repo.On("CreateConversation", mock.Anything, mock.AnythingOfType("*domain.Conversation")).
		Return(nil).
		Run(func(args mock.Arguments) {
			conv := args.Get(1).(*model.Conversation) // now safe: arg1 is the conversation
			conv.ID = 777
			conv.CreatedAt = time.Now()
		})

	// AddParticipant called twice
	repo.On("AddParticipant", mock.Anything, mock.AnythingOfType("*domain.Participant")).
		Return(nil).Times(2)

	// SaveMessage
	repo.On("SaveMessage", mock.Anything, mock.AnythingOfType("*domain.Message")).
		Return(nil)

	msg, err := svc.SendMessage(ctx, sender, recipient, content)

	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, int64(777), msg.ConversationID)
	assert.Equal(t, content, msg.Content)

	repo.AssertExpectations(t)
}

func TestSendMessage_SaveMessageError(t *testing.T) {
	ctx := context.Background()

	repo := new(MockRepo)
	wsMgr := websocket.NewClientManager()
	svc := NewChatService(repo, wsMgr)

	sender := int64(1)
	recipient := int64(2)

	conv := &model.Conversation{ID: 10}

	repo.On("FindOneToOneConversation", mock.Anything, sender, recipient).
		Return(conv, nil)

	repo.On("SaveMessage", mock.Anything, mock.Anything).
		Return(errors.New("db error"))

	msg, err := svc.SendMessage(ctx, sender, recipient, "x")

	assert.Nil(t, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")

	repo.AssertExpectations(t)
}
