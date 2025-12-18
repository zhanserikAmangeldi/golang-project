package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	grpcMocks "github.com/zhanserikAmangeldi/chat-service/internal/adapters/grpc/mocks"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
	repoMocks "github.com/zhanserikAmangeldi/chat-service/internal/core/ports/mocks"
	redisMocks "github.com/zhanserikAmangeldi/chat-service/internal/redis/mocks"
)

func TestSendMessage_NewConversation(t *testing.T) {
	mockRepo := new(repoMocks.MockChatRepository)
	mockRedis := new(redisMocks.MockRedisClient)
	mockUserClient := new(grpcMocks.MockUserClient)

	service := NewChatService(mockRepo, mockRedis, mockUserClient)

	ctx := context.Background()
	senderID := int64(1)
	recipientID := int64(2)
	content := "Hello, World!"

	mockRepo.On("FindOneToOneConversation", ctx, senderID, recipientID).
		Return(nil, nil)

	mockUserClient.On("ValidateUserExists", ctx, recipientID).
		Return(true, nil)

	mockRepo.On("CreateConversation", ctx, mock.AnythingOfType("*model.Conversation")).
		Return(nil)

	mockRepo.On("AddParticipant", ctx, mock.AnythingOfType("*model.Participant")).
		Return(nil).Twice()

	mockRepo.On("SaveMessage", ctx, mock.AnythingOfType("*model.Message")).
		Return(nil)

	mockRepo.On("GetParticipants", ctx, int64(1)).
		Return([]int64{senderID, recipientID}, nil)

	mockRedis.On("Publish", ctx, mock.AnythingOfType("model.Message"), mock.AnythingOfType("[]int64")).
		Return(nil)

	message, err := service.SendMessage(ctx, senderID, recipientID, content, 0, "text", nil, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, int64(42), message.ID) // Mock sets ID to 42
	assert.Equal(t, senderID, message.SenderID)
	assert.Equal(t, content, message.Content)
	assert.Equal(t, "text", message.MessageType)

	mockRepo.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestSendMessage_ExistingConversation(t *testing.T) {
	mockRepo := new(repoMocks.MockChatRepository)
	mockRedis := new(redisMocks.MockRedisClient)
	mockUserClient := new(grpcMocks.MockUserClient)

	service := NewChatService(mockRepo, mockRedis, mockUserClient)

	ctx := context.Background()
	senderID := int64(1)
	conversationID := int64(5)
	content := "Hello again!"

	existingConv := &model.Conversation{
		ID:        conversationID,
		IsGroup:   false,
		CreatedAt: time.Now(),
	}

	mockRepo.On("GetConversationByID", ctx, conversationID).
		Return(existingConv, nil)

	mockRepo.On("SaveMessage", ctx, mock.AnythingOfType("*model.Message")).
		Return(nil)

	mockRepo.On("GetParticipants", ctx, conversationID).
		Return([]int64{senderID, int64(2)}, nil)

	mockRedis.On("Publish", ctx, mock.AnythingOfType("model.Message"), mock.AnythingOfType("[]int64")).
		Return(nil)

	message, err := service.SendMessage(ctx, senderID, 0, content, conversationID, "text", nil, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, conversationID, message.ConversationID)
	assert.Equal(t, content, message.Content)

	mockRepo.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestCreateGroup_Success(t *testing.T) {
	mockRepo := new(repoMocks.MockChatRepository)
	mockRedis := new(redisMocks.MockRedisClient)
	mockUserClient := new(grpcMocks.MockUserClient)

	service := NewChatService(mockRepo, mockRedis, mockUserClient)

	ctx := context.Background()
	creatorID := int64(1)
	memberIDs := []int64{2, 3, 4}
	groupName := "Test Group"

	mockUserClient.On("ValidateUsersExist", ctx, []int64{2, 3, 4, 1}).
		Return(true, nil)

	mockRepo.On("CreateConversation", ctx, mock.AnythingOfType("*model.Conversation")).
		Return(nil)

	mockRepo.On("AddParticipant", ctx, mock.AnythingOfType("*model.Participant")).
		Return(nil).Times(4)

	conversation, err := service.CreateGroup(ctx, groupName, creatorID, memberIDs)

	assert.NoError(t, err)
	assert.NotNil(t, conversation)
	assert.True(t, conversation.IsGroup)
	assert.Equal(t, groupName, conversation.Name)

	mockRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestCreateGroup_UserNotExist(t *testing.T) {
	mockRepo := new(repoMocks.MockChatRepository)
	mockRedis := new(redisMocks.MockRedisClient)
	mockUserClient := new(grpcMocks.MockUserClient)

	service := NewChatService(mockRepo, mockRedis, mockUserClient)

	ctx := context.Background()
	creatorID := int64(1)
	memberIDs := []int64{2, 999} // 999 doesn't exist
	groupName := "Test Group"

	mockUserClient.On("ValidateUsersExist", ctx, []int64{2, 999, 1}).
		Return(false, nil)

	conversation, err := service.CreateGroup(ctx, groupName, creatorID, memberIDs)

	assert.Error(t, err)
	assert.Nil(t, conversation)
	assert.Contains(t, err.Error(), "do not exist")

	mockUserClient.AssertExpectations(t)
}

func TestMarkMessageAsRead_Success(t *testing.T) {
	mockRepo := new(repoMocks.MockChatRepository)
	mockRedis := new(redisMocks.MockRedisClient)
	mockUserClient := new(grpcMocks.MockUserClient)

	service := NewChatService(mockRepo, mockRedis, mockUserClient)

	ctx := context.Background()
	messageID := int64(42)
	userID := int64(2)
	senderID := int64(1)

	message := &model.Message{
		ID:             messageID,
		ConversationID: 5,
		SenderID:       senderID,
		Content:        "Test message",
	}

	mockRepo.On("GetMessageByID", ctx, messageID).
		Return(message, nil)

	mockRepo.On("IsParticipant", ctx, int64(5), userID).
		Return(true, nil)

	mockRepo.On("MarkMessageAsRead", ctx, messageID, userID).
		Return(nil)

	mockRepo.On("GetParticipants", ctx, int64(5)).
		Return([]int64{senderID, userID}, nil)

	mockRedis.On("PublishReadReceipt", ctx, mock.AnythingOfType("model.MessageRead"), []int64{senderID}).
		Return(nil)

	err := service.MarkMessageAsRead(ctx, messageID, userID)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestMarkMessageAsRead_OwnMessage(t *testing.T) {
	mockRepo := new(repoMocks.MockChatRepository)
	mockRedis := new(redisMocks.MockRedisClient)
	mockUserClient := new(grpcMocks.MockUserClient)

	service := NewChatService(mockRepo, mockRedis, mockUserClient)

	ctx := context.Background()
	messageID := int64(42)
	userID := int64(1)

	message := &model.Message{
		ID:             messageID,
		ConversationID: 5,
		SenderID:       userID, // Same as reader
		Content:        "Test message",
	}

	mockRepo.On("GetMessageByID", ctx, messageID).
		Return(message, nil)

	err := service.MarkMessageAsRead(ctx, messageID, userID)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockRedis.AssertNotCalled(t, "PublishReadReceipt")
}

func TestAddReaction_Success(t *testing.T) {
	mockRepo := new(repoMocks.MockChatRepository)
	mockRedis := new(redisMocks.MockRedisClient)
	mockUserClient := new(grpcMocks.MockUserClient)

	service := NewChatService(mockRepo, mockRedis, mockUserClient)

	ctx := context.Background()
	messageID := int64(42)
	userID := int64(2)
	reaction := "👍"

	message := &model.Message{
		ID:             messageID,
		ConversationID: 5,
		SenderID:       1,
	}

	mockRepo.On("GetMessageByID", ctx, messageID).
		Return(message, nil)

	mockRepo.On("IsParticipant", ctx, int64(5), userID).
		Return(true, nil)

	mockRepo.On("AddReaction", ctx, messageID, userID, reaction).
		Return(nil)

	mockRepo.On("GetParticipants", ctx, int64(5)).
		Return([]int64{int64(1), userID}, nil)

	mockRedis.On("PublishReaction", ctx, mock.AnythingOfType("model.Reaction"), mock.AnythingOfType("[]int64")).
		Return(nil)

	err := service.AddReaction(ctx, messageID, userID, reaction)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestEditMessage_Success(t *testing.T) {
	mockRepo := new(repoMocks.MockChatRepository)
	mockRedis := new(redisMocks.MockRedisClient)
	mockUserClient := new(grpcMocks.MockUserClient)

	service := NewChatService(mockRepo, mockRedis, mockUserClient)

	ctx := context.Background()
	messageID := int64(42)
	userID := int64(1)
	newContent := "Edited message"

	message := &model.Message{
		ID:             messageID,
		ConversationID: 5,
		SenderID:       userID,
		Content:        "Original message",
		DeletedAt:      nil,
	}

	editedMessage := &model.Message{
		ID:             messageID,
		ConversationID: 5,
		SenderID:       userID,
		Content:        newContent,
		EditedAt:       &time.Time{},
	}

	mockRepo.On("GetMessageByID", ctx, messageID).
		Return(message, nil).Once()

	mockRepo.On("EditMessage", ctx, messageID, newContent).
		Return(nil)

	mockRepo.On("GetMessageByID", ctx, messageID).
		Return(editedMessage, nil).Once()

	mockRepo.On("GetParticipants", ctx, int64(5)).
		Return([]int64{userID, int64(2)}, nil)

	mockRedis.On("PublishMessageEdit", ctx, mock.AnythingOfType("model.Message"), mock.AnythingOfType("[]int64")).
		Return(nil)

	err := service.EditMessage(ctx, messageID, userID, newContent)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestEditMessage_NotOwner(t *testing.T) {
	mockRepo := new(repoMocks.MockChatRepository)
	mockRedis := new(redisMocks.MockRedisClient)
	mockUserClient := new(grpcMocks.MockUserClient)

	service := NewChatService(mockRepo, mockRedis, mockUserClient)

	ctx := context.Background()
	messageID := int64(42)
	userID := int64(2) // Not the sender
	newContent := "Edited message"

	message := &model.Message{
		ID:             messageID,
		ConversationID: 5,
		SenderID:       int64(1), // Different user
		Content:        "Original message",
	}

	mockRepo.On("GetMessageByID", ctx, messageID).
		Return(message, nil)

	err := service.EditMessage(ctx, messageID, userID, newContent)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only message sender")

	mockRepo.AssertExpectations(t)
}

func TestDeleteMessage_Success(t *testing.T) {
	mockRepo := new(repoMocks.MockChatRepository)
	mockRedis := new(redisMocks.MockRedisClient)
	mockUserClient := new(grpcMocks.MockUserClient)

	service := NewChatService(mockRepo, mockRedis, mockUserClient)

	ctx := context.Background()
	messageID := int64(42)
	userID := int64(1)

	message := &model.Message{
		ID:             messageID,
		ConversationID: 5,
		SenderID:       userID,
	}

	mockRepo.On("GetMessageByID", ctx, messageID).
		Return(message, nil)

	mockRepo.On("DeleteMessage", ctx, messageID).
		Return(nil)

	mockRepo.On("GetParticipants", ctx, int64(5)).
		Return([]int64{userID, int64(2)}, nil)

	mockRedis.On("PublishMessageDeletion", ctx, messageID, mock.AnythingOfType("[]int64")).
		Return(nil)

	err := service.DeleteMessage(ctx, messageID, userID)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}
