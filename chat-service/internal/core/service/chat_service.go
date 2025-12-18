package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/zhanserikAmangeldi/chat-service/internal/adapters/grpc"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/ports"
	redisAdapter "github.com/zhanserikAmangeldi/chat-service/internal/redis"
)

type ChatService struct {
	repo       ports.ChatRepository
	redis      redisAdapter.IRedisClient
	userClient grpc.IUserClient
}

func NewChatService(repo ports.ChatRepository, redis redisAdapter.IRedisClient, userClient grpc.IUserClient) *ChatService {
	return &ChatService{
		repo:       repo,
		redis:      redis,
		userClient: userClient,
	}
}

func (s *ChatService) CreateGroup(ctx context.Context, name string, creatorID int64, memberIDs []int64) (*model.Conversation, error) {
	allMembers := append(memberIDs, creatorID)
	log.Printf("name: %v, creatorID: %v, allMembers: %v", name, creatorID, allMembers)
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

func (s *ChatService) SendMessage(ctx context.Context, senderID, recipientID int64, content string, conversationID int64, messageType string, fileURL, fileName, mimeType *string, fileSize *int64) (*model.Message, error) {
	var conv *model.Conversation
	var err error

	if conversationID > 0 {
		conv, err = s.repo.GetConversationByID(ctx, conversationID)
		if err != nil {
			return nil, err
		}
		if conv == nil {
			return nil, errors.New("conversation not found")
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

	if messageType == "" {
		messageType = "text"
	}

	msg := &model.Message{
		ConversationID: conv.ID,
		SenderID:       senderID,
		Content:        content,
		MessageType:    messageType,
		FileURL:        fileURL,
		FileName:       fileName,
		FileSize:       fileSize,
		MimeType:       mimeType,
		CreatedAt:      time.Now(),
	}

	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	participantIDs, err := s.repo.GetParticipants(ctx, conv.ID)
	if err != nil {
		return msg, nil
	}

	recipients := make([]int64, 0)
	for _, pid := range participantIDs {
		if pid != senderID {
			recipients = append(recipients, pid)
		}
	}

	if len(recipients) > 0 {
		_ = s.redis.Publish(ctx, *msg, recipients)
	}

	return msg, nil
}

func (s *ChatService) GetHistory(ctx context.Context, conversationID int64, limit, offset int) ([]model.Message, error) {
	if limit == 0 {
		limit = 50
	}
	return s.repo.GetMessages(ctx, conversationID, limit, offset)
}

func (s *ChatService) GetUserConversations(ctx context.Context, userID int64, limit, offset int) ([]model.ConversationWithLastMessage, error) {
	if limit == 0 {
		limit = 20
	}
	return s.repo.GetUserConversations(ctx, userID, limit, offset)
}

func (s *ChatService) MarkMessageAsRead(ctx context.Context, messageID, userID int64) error {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return err
	}

	if msg.SenderID == userID {
		return nil
	}

	isParticipant, err := s.repo.IsParticipant(ctx, msg.ConversationID, userID)
	if err != nil || !isParticipant {
		return errors.New("user is not a participant of this conversation")
	}

	err = s.repo.MarkMessageAsRead(ctx, messageID, userID)
	if err != nil {
		return err
	}

	readReceipt := model.MessageRead{
		MessageID: messageID,
		UserID:    userID,
		ReadAt:    time.Now(),
	}

	_, _ = s.repo.GetParticipants(ctx, msg.ConversationID)

	recipients := []int64{msg.SenderID}

	_ = s.redis.PublishReadReceipt(ctx, readReceipt, recipients)

	return nil
}

func (s *ChatService) AddReaction(ctx context.Context, messageID, userID int64, reaction string) error {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return err
	}

	isParticipant, err := s.repo.IsParticipant(ctx, msg.ConversationID, userID)
	if err != nil || !isParticipant {
		return errors.New("user is not a participant of this conversation")
	}

	err = s.repo.AddReaction(ctx, messageID, userID, reaction)
	if err != nil {
		return err
	}

	reactionEvent := model.Reaction{
		MessageID: messageID,
		UserID:    userID,
		Reaction:  reaction,
		CreatedAt: time.Now(),
	}

	participants, _ := s.repo.GetParticipants(ctx, msg.ConversationID)
	_ = s.redis.PublishReaction(ctx, reactionEvent, participants)

	return nil
}

func (s *ChatService) RemoveReaction(ctx context.Context, messageID, userID int64, reaction string) error {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return err
	}

	err = s.repo.RemoveReaction(ctx, messageID, userID, reaction)
	if err != nil {
		return err
	}

	reactionEvent := model.Reaction{
		MessageID: messageID,
		UserID:    userID,
		Reaction:  reaction,
	}

	participants, _ := s.repo.GetParticipants(ctx, msg.ConversationID)
	_ = s.redis.PublishReactionRemoval(ctx, reactionEvent, participants)

	return nil
}

func (s *ChatService) EditMessage(ctx context.Context, messageID, userID int64, newContent string) error {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return err
	}

	if msg.SenderID != userID {
		return errors.New("only message sender can edit the message")
	}

	if msg.DeletedAt != nil {
		return errors.New("cannot edit deleted message")
	}

	err = s.repo.EditMessage(ctx, messageID, newContent)
	if err != nil {
		return err
	}

	updatedMsg, _ := s.repo.GetMessageByID(ctx, messageID)
	if updatedMsg != nil {
		participants, _ := s.repo.GetParticipants(ctx, msg.ConversationID)
		_ = s.redis.PublishMessageEdit(ctx, *updatedMsg, participants)
	}

	return nil
}

func (s *ChatService) DeleteMessage(ctx context.Context, messageID, userID int64) error {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return err
	}

	if msg.SenderID != userID {
		return errors.New("only message sender can delete the message")
	}

	err = s.repo.DeleteMessage(ctx, messageID)
	if err != nil {
		return err
	}

	participants, _ := s.repo.GetParticipants(ctx, msg.ConversationID)
	_ = s.redis.PublishMessageDeletion(ctx, messageID, participants)

	return nil
}
