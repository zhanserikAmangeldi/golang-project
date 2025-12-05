package service

import (
	"context"
	"encoding/json"
	"time"
	
	userpb "github.com/zhanserikAmangeldi/proto/userpb"
	"chat-service/internal/adapters/websocket"
	"chat-service/internal/core/domain"
	"chat-service/internal/core/ports"
)

type ChatService struct {
	repo       ports.ChatRepository
	wsManager  *websocket.ClientManager
	userClient userpb.UserServiceClient
}

func NewChatService(repo ports.ChatRepository, wsManager *websocket.ClientManager, userClient userpb.UserServiceClient) *ChatService {
	return &ChatService{
		repo:       repo,
		wsManager:  wsManager,
		userClient: userClient,  // ← ДОБАВЛЕНА ЗАПЯТАЯ
	}
}

// SendMessage handles the logic:
// 1. Check if conversation exists (if not, create it)
// 2. Save message
func (s *ChatService) SendMessage(ctx context.Context, senderID, recipientID int64, content string) (*domain.Message, error) {
	// 1. Validate users exist - ИСПРАВЛЕНО: GetUserById вместо GetUser, правильный тип запроса
	if _, err := s.userClient.GetUserById(ctx, &userpb.GetUserRequest{Id: senderID}); err != nil {
		return nil, err
	}
	if _, err := s.userClient.GetUserById(ctx, &userpb.GetUserRequest{Id: recipientID}); err != nil {
		return nil, err
	}

	conv, err := s.repo.FindOneToOneConversation(ctx, senderID, recipientID)
	if err != nil {
		return nil, err
	}

	// 2. If no conversation, create one
	if conv == nil {
		newConv := &domain.Conversation{
			IsGroup:   false,
			CreatedAt: time.Now(),
		}
		// Repo will fill in the ID
		if err := s.repo.CreateConversation(ctx, newConv); err != nil {
			return nil, err
		}
		conv = newConv

		// Add participants
		s.repo.AddParticipant(ctx, &domain.Participant{
			ConversationID: conv.ID,
			UserID:         senderID,
			JoinedAt:       time.Now(),
		})
		s.repo.AddParticipant(ctx, &domain.Participant{
			ConversationID: conv.ID,
			UserID:         recipientID,
			JoinedAt:       time.Now(),
		})
	}

	// 3. Create the Message object
	msg := &domain.Message{
		ConversationID: conv.ID,
		SenderID:       senderID,
		Content:        content,
		CreatedAt:      time.Now(),
	}

	// 4. Save to DB
	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	// 5. Send via WebSocket if recipient is online
	if conn, ok := s.wsManager.GetClient(recipientID); ok {
		// Marshal message to JSON
		msgBytes, _ := json.Marshal(msg)
		// Send to Recipient
		conn.WriteMessage(1, msgBytes)
	}

	return msg, nil
}

func (s *ChatService) GetHistory(ctx context.Context, conversationID int64) ([]domain.Message, error) {
	return s.repo.GetMessages(ctx, conversationID, 50, 0) // Limit 50 for now
}