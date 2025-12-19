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

func (s *ChatService) CreateGroup(ctx context.Context, name string, creatorID int64, memberIDs []int64) (*model.Conversation, error) {
	allMembers := append(memberIDs, creatorID)
	exists, err := s.userClient.ValidateUsersExist(ctx, allMembers)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("one or more users do not exist")
	}

	conv := &model.Conversation{
		IsGroup:   true,
		Name:      name,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}

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

// SendMessage ТЕПЕРЬ С 10 АРГУМЕНТАМИ
func (s *ChatService) SendMessage(
	ctx context.Context,
	senderID, recipientID int64,
	content string,
	conversationID int64,
	messageType string, // 6-й
	fileURL *string,    // 7-й
	fileName *string,   // 8-й
	mimeType *string,   // 9-й
	fileSize *int64,    // 10-й
) (*model.Message, error) {
	var conv *model.Conversation
	var err error

	if conversationID > 0 {
		conv, err = s.repo.GetConversationByID(ctx, conversationID)
		if err != nil {
			return nil, err
		}
	} else {
		conv, err = s.repo.FindOneToOneConversation(ctx, senderID, recipientID)
		if err != nil {
			return nil, err
		}

		if conv == nil {
			exists, _ := s.userClient.ValidateUserExists(ctx, recipientID)
			if !exists {
				return nil, errors.New("recipient user does not exist")
			}

			newConv := &model.Conversation{
				IsGroup:   false,
				CreatedAt: time.Now(),
			}

			if err := s.repo.CreateConversation(ctx, newConv); err != nil {
				return nil, err
			}
			conv = newConv

			s.repo.AddParticipant(ctx, &model.Participant{ConversationID: conv.ID, UserID: senderID, JoinedAt: time.Now()})
			s.repo.AddParticipant(ctx, &model.Participant{ConversationID: conv.ID, UserID: recipientID, JoinedAt: time.Now()})
		}
	}

	if conv == nil {
		return nil, errors.New("conversation not found")
	}

	// Создаем объект сообщения со всеми 10 параметрами
	msg := &model.Message{
		ConversationID: conv.ID,
		SenderID:       senderID,
		Content:        content,
		MessageType:    messageType,
		FileURL:        fileURL,
		FileName:       fileName,
		MimeType:       mimeType,
		FileSize:       fileSize,
		CreatedAt:      time.Now(),
	}

	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	participantIDs, err := s.repo.GetParticipants(ctx, conv.ID)
	if err == nil {
		recipients := make([]int64, 0)
		for _, pid := range participantIDs {
			if pid != senderID {
				recipients = append(recipients, pid)
			}
		}
		if len(recipients) > 0 {
			_ = s.redis.Publish(ctx, *msg, recipients)
		}
	}

	return msg, nil
}

func (s *ChatService) GetHistory(ctx context.Context, conversationID int64) ([]model.Message, error) {
	return s.repo.GetMessages(ctx, conversationID, 50, 0)
}