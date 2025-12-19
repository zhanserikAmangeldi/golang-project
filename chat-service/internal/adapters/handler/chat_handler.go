package handler

import (
	"net/http"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/service"
)

type ChatHandler struct {
	service *service.ChatService
}

func NewChatHandler(s *service.ChatService) *ChatHandler {
	return &ChatHandler{service: s}
}

// GetHistory
// @Summary Получить историю сообщений
// @Description Возвращает список сообщений для конкретного чата
// @Tags messages
// @Security BearerAuth
// @Param conversation_id query int true "ID чата"
// @Success 200 {array} model.Message
// @Router /api/v1/messages/history [get]
func (h *ChatHandler) GetHistory(w http.ResponseWriter, r *http.Request) {}

// GetConversations
// @Summary Список диалогов
// @Description Получить список всех активных диалогов пользователя
// @Tags conversations
// @Security BearerAuth
// @Success 200 {array} model.Conversation
// @Router /api/v1/conversations [get]
func (h *ChatHandler) GetConversations(w http.ResponseWriter, r *http.Request) {}

// MarkAsRead
// @Summary Прочитать сообщения
// @Description Пометить сообщения в чате как прочитанные
// @Tags messages
// @Security BearerAuth
// @Param conversation_id body int true "ID чата"
// @Success 200 {object} map[string]string
// @Router /api/v1/messages/read [post]
func (h *ChatHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {}

// AddReaction
// @Summary Добавить реакцию
// @Description Добавить эмодзи к сообщению
// @Tags reactions
// @Security BearerAuth
// @Param message_id body int true "ID сообщения"
// @Param reaction body string true "Эмодзи"
// @Success 200 {object} map[string]string
// @Router /api/v1/messages/reactions/add [post]
func (h *ChatHandler) AddReaction(w http.ResponseWriter, r *http.Request) {}

// RemoveReaction
// @Summary Удалить реакцию
// @Tags reactions
// @Security BearerAuth
// @Router /api/v1/messages/reactions/remove [post]
func (h *ChatHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {}

// EditMessage
// @Summary Редактировать сообщение
// @Tags messages
// @Security BearerAuth
// @Router /api/v1/messages/edit [put]
func (h *ChatHandler) EditMessage(w http.ResponseWriter, r *http.Request) {}

// DeleteMessage
// @Summary Удалить сообщение
// @Tags messages
// @Security BearerAuth
// @Router /api/v1/messages/delete [delete]
func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {}