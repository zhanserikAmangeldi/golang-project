package handler

import (
	"net/http"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/service"
)

type FileHandler struct {
	minio *service.MinioService 
	chat  *service.ChatService
}

func NewFileHandler(m *service.MinioService, c *service.ChatService) *FileHandler {
	return &FileHandler{minio: m, chat: c}
}

// UploadFile
// @Summary Загрузить файл
// @Description Загружает файл в хранилище MinIO
// @Tags files
// @Accept multipart/form-data
// @Security BearerAuth
// @Param file formData file true "Файл для загрузки"
// @Success 200 {object} map[string]string
// @Router /api/v1/files/upload [post]
func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {}

// SendMessageWithFile
// @Summary Отправить сообщение с файлом
// @Tags files
// @Security BearerAuth
// @Router /api/v1/files/send [post]
func (h *FileHandler) SendMessageWithFile(w http.ResponseWriter, r *http.Request) {}

// GetFile
// @Summary Получить ссылку на файл
// @Tags files
// @Security BearerAuth
// @Param file_name query string true "Имя файла"
// @Router /api/v1/files/get [get]
func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {}