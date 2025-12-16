package service

import (
	"context"
	"errors"
	"time"

	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/grpc"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/ports"
	redisAdapter "github.com/zhanserikAmangeldi/chat-service/internal/redis"
)

type ChatService struct {
	repo       ports.ChatRepository
	redis      *redisAdapter.RedisClient
	userClient *grpc.UserClient
}

func NewChatService(repo ports.ChatRepository, redis *redisAdapter.RedisClient, userClient *grpc.UserClient) *ChatService {
	return &ChatService{
		repo:       repo,
		redis:      redis,
		userClient: userClient,
	}
}

// CreateGroup validates users via gRPC and creates a named conversation
func (s *ChatService) CreateGroup(ctx context.Context, name string, creatorID int64, memberIDs []int64) (*model.Conversation, error) {
	// 1. Validate all users exist via gRPC
	allMembers := append(memberIDs, creatorID)
	exists, err := s.userClient.ValidateUsersExist(ctx, allMembers)
	if err != nil {
		return nil, err // gRPC error
	}
	if !exists {
		return nil, errors.New("one or more users do not exist")
	}

	// 2. Create Conversation
	conv := &model.Conversation{
		IsGroup:   true,
		Name:      name,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}

	// 3. Add Participants
	for _, uid := range allMembers {
		p := &model.Participant{
			ConversationID: conv.ID,
			UserID:         uid,
			JoinedAt:       time.Now(),
		}
		s.repo.AddParticipant(ctx, p)
	}

	return conv, nil
}

// SendMessage handles logic for both 1:1 and Group chats.
//
// Scenario A (New 1:1): Pass conversationID=0. We look up or create the chat.
// Scenario B (Existing): Pass conversationID=N. We send directly to that room.
func (s *ChatService) SendMessage(ctx context.Context, senderID, recipientID int64, content string, conversationID int64) (*model.Message, error) {
	var conv *model.Conversation
	var err error

	// CASE 1: Existing Conversation ID provided (Group or existing 1:1)
	if conversationID > 0 {
		conv, err = s.repo.GetConversationByID(ctx, conversationID)
		if err != nil {
			return nil, err
		}
		if conv == nil {
			return nil, errors.New("conversation not found")
		}
	} else {
		// CASE 2: New 1:1 Chat (No Conversation ID yet)

		// A. Check if 1:1 already exists in DB
		conv, err = s.repo.FindOneToOneConversation(ctx, senderID, recipientID)
		if err != nil {
			return nil, err
		}

		// B. If not found, create it
		if conv == nil {
			// Validate recipient exists via gRPC
			exists, _ := s.userClient.ValidateUserExists(ctx, recipientID)
			if !exists {
				return nil, errors.New("recipient user does not exist")
			}

			newConv := &model.Conversation{
				IsGroup:   false,
				CreatedAt: time.Now(),
			}

			// Repository generates and sets the ID
			if err := s.repo.CreateConversation(ctx, newConv); err != nil {
				return nil, err
			}
			conv = newConv

			// Add both participants
			s.repo.AddParticipant(ctx, &model.Participant{ConversationID: conv.ID, UserID: senderID, JoinedAt: time.Now()})
			s.repo.AddParticipant(ctx, &model.Participant{ConversationID: conv.ID, UserID: recipientID, JoinedAt: time.Now()})
		}
	}

	// 3. Create the Message object
	msg := &model.Message{
		ConversationID: conv.ID,
		SenderID:       senderID,
		Content:        content,
		CreatedAt:      time.Now(),
	}

	// 4. Save to DB
	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	// 5. Broadcast via Redis

	// Fetch who is in this room
	participantIDs, err := s.repo.GetParticipants(ctx, conv.ID)
	if err != nil {
		// Log error, but don't fail because message is safely in DB
		return msg, nil
	}

	// Filter out the sender (optional)
	recipients := make([]int64, 0)
	for _, pid := range participantIDs {
		if pid != senderID {
			recipients = append(recipients, pid)
		}
	}

	// Publish to Redis if there are recipients
	if len(recipients) > 0 {
		// This pushes the message to the Redis Channel.
		// The background listener in main.go will pick this up and send to WebSockets.
		_ = s.redis.Publish(ctx, *msg, recipients)
	}

	return msg, nil
}

func (s *ChatService) GetHistory(ctx context.Context, conversationID int64) ([]model.Message, error) {
	return s.repo.GetMessages(ctx, conversationID, 50, 0)
}
