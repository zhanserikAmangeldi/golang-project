package service

import (
	"github.com/zhanserikAmangeldi/chat-service/internal/repository"
)

type MessageService struct {
	repo *repository.MessageRepo
}

func NewMessageService(repo *repository.MessageRepo) *MessageService {
	return &MessageService{repo: repo}
}

func (s *MessageService) CreateMessage(convID, senderID, content, msgType string) (*repository.Message, error) {
	// Business rules can be added here (e.g., content length)
	if msgType == "" {
		msgType = "text"
	}
	return s.repo.Save(convID, senderID, content, msgType)
}
