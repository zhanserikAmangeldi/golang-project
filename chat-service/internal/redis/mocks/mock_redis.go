package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
	"github.com/zhanserikAmangeldi/chat-service/internal/redis"
)

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Publish(ctx context.Context, msg model.Message, recipients []int64) error {
	args := m.Called(ctx, msg, recipients)
	return args.Error(0)
}

func (m *MockRedisClient) PublishTyping(ctx context.Context, event model.TypingEvent, recipients []int64) error {
	args := m.Called(ctx, event, recipients)
	return args.Error(0)
}

func (m *MockRedisClient) PublishStatus(ctx context.Context, event model.OnlineStatusEvent, recipients []int64) error {
	args := m.Called(ctx, event, recipients)
	return args.Error(0)
}

func (m *MockRedisClient) PublishReaction(ctx context.Context, reaction model.Reaction, recipients []int64) error {
	args := m.Called(ctx, reaction, recipients)
	return args.Error(0)
}

func (m *MockRedisClient) PublishReactionRemoval(ctx context.Context, reaction model.Reaction, recipients []int64) error {
	args := m.Called(ctx, reaction, recipients)
	return args.Error(0)
}

func (m *MockRedisClient) PublishReadReceipt(ctx context.Context, readReceipt model.MessageRead, recipients []int64) error {
	args := m.Called(ctx, readReceipt, recipients)
	return args.Error(0)
}

func (m *MockRedisClient) PublishMessageEdit(ctx context.Context, msg model.Message, recipients []int64) error {
	args := m.Called(ctx, msg, recipients)
	return args.Error(0)
}

func (m *MockRedisClient) PublishMessageDeletion(ctx context.Context, messageID int64, recipients []int64) error {
	args := m.Called(ctx, messageID, recipients)
	return args.Error(0)
}

func (m *MockRedisClient) Subscribe(ctx context.Context) <-chan redis.BroadcastMessage {
	args := m.Called(ctx)
	return args.Get(0).(<-chan redis.BroadcastMessage)
}
