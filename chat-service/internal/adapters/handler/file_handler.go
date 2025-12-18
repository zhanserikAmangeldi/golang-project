package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/zhanserikAmangeldi/chat-service/internal/core/service"
	"github.com/zhanserikAmangeldi/chat-service/internal/middleware"
)

type FileHandler struct {
	minioService *service.MinioService
	chatService  *service.ChatService
}

func NewFileHandler(minioService *service.MinioService, chatService *service.ChatService) *FileHandler {
	return &FileHandler{
		minioService: minioService,
		chatService:  chatService,
	}
}

const (
	MaxImageSize = 10 * 1024 * 1024  // 10 MB
	MaxFileSize  = 50 * 1024 * 1024  // 50 MB
	MaxAudioSize = 20 * 1024 * 1024  // 20 MB
	MaxVideoSize = 100 * 1024 * 1024 // 100 MB
)

type UploadFileResponse struct {
	FileURL     string `json:"file_url"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	MimeType    string `json:"mime_type"`
	MessageType string `json:"message_type"`
	ObjectName  string `json:"object_name"`
	Bucket      string `json:"bucket"`
}

func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileSize := header.Size
	if err := validateFileSize(contentType, fileSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isAllowedFileType(contentType) {
		http.Error(w, "File type not allowed: "+contentType, http.StatusBadRequest)
		return
	}

	bucket := service.DetermineBucket(contentType)

	extension := filepath.Ext(header.Filename)
	objectName := fmt.Sprintf("%d/%s%s", userID, generateRandomID(), extension)

	err = h.minioService.UploadFile(r.Context(), bucket, objectName, file, fileSize, contentType)
	if err != nil {
		http.Error(w, "Failed to upload file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fileURL, err := h.minioService.GetFileURL(r.Context(), bucket, objectName, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate file URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	messageType := determineMessageType(contentType)

	response := UploadFileResponse{
		FileURL:     fileURL,
		FileName:    header.Filename,
		FileSize:    fileSize,
		MimeType:    contentType,
		MessageType: messageType,
		ObjectName:  objectName,
		Bucket:      bucket,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *FileHandler) SendMessageWithFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	conversationIDStr := r.FormValue("conversation_id")
	recipientIDStr := r.FormValue("recipient_id")
	caption := r.FormValue("caption")

	var conversationID, recipientID int64
	if conversationIDStr != "" {
		conversationID, _ = strconv.ParseInt(conversationIDStr, 10, 64)
	}
	if recipientIDStr != "" {
		recipientID, _ = strconv.ParseInt(recipientIDStr, 10, 64)
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileSize := header.Size
	if err := validateFileSize(contentType, fileSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isAllowedFileType(contentType) {
		http.Error(w, "File type not allowed", http.StatusBadRequest)
		return
	}

	bucket := service.DetermineBucket(contentType)
	extension := filepath.Ext(header.Filename)
	objectName := fmt.Sprintf("%d/%s%s", userID, generateRandomID(), extension)

	err = h.minioService.UploadFile(r.Context(), bucket, objectName, file, fileSize, contentType)
	if err != nil {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	fileURL, err := h.minioService.GetFileURL(r.Context(), bucket, objectName, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate file URL", http.StatusInternalServerError)
		return
	}

	messageType := determineMessageType(contentType)

	fileName := header.Filename
	msg, err := h.chatService.SendMessage(
		r.Context(),
		userID,
		recipientID,
		caption,
		conversationID,
		messageType,
		&fileURL,
		&fileName,
		&contentType,
		&fileSize,
	)
	if err != nil {
		_ = h.minioService.DeleteFile(r.Context(), bucket, objectName)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	bucket := r.URL.Query().Get("bucket")
	objectName := r.URL.Query().Get("object_name")

	if bucket == "" || objectName == "" {
		http.Error(w, "bucket and object_name are required", http.StatusBadRequest)
		return
	}

	_, err := h.minioService.GetFileInfo(r.Context(), bucket, objectName)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	fileURL, err := h.minioService.GetFileURL(r.Context(), bucket, objectName, time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate file URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"file_url": fileURL,
	})
}

func validateFileSize(contentType string, size int64) error {
	switch {
	case isImage(contentType):
		if size > MaxImageSize {
			return fmt.Errorf("image size exceeds limit of %d MB", MaxImageSize/(1024*1024))
		}
	case isAudio(contentType):
		if size > MaxAudioSize {
			return fmt.Errorf("audio size exceeds limit of %d MB", MaxAudioSize/(1024*1024))
		}
	case isVideo(contentType):
		if size > MaxVideoSize {
			return fmt.Errorf("video size exceeds limit of %d MB", MaxVideoSize/(1024*1024))
		}
	default:
		if size > MaxFileSize {
			return fmt.Errorf("file size exceeds limit of %d MB", MaxFileSize/(1024*1024))
		}
	}
	return nil
}

func isAllowedFileType(contentType string) bool {
	allowed := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp",
		"application/pdf", "application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain", "text/csv",
		"application/zip", "application/x-rar-compressed",
		"audio/mpeg", "audio/mp3", "audio/wav", "audio/ogg", "audio/webm",
		"video/mp4", "video/mpeg", "video/quicktime", "video/webm",
	}

	for _, allowed := range allowed {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func determineMessageType(contentType string) string {
	switch {
	case isImage(contentType):
		return "image"
	case isAudio(contentType):
		return "audio"
	case isVideo(contentType):
		return "video"
	default:
		return "file"
	}
}

func isImage(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

func isAudio(mimeType string) bool {
	return strings.HasPrefix(mimeType, "audio/")
}

func isVideo(mimeType string) bool {
	return strings.HasPrefix(mimeType, "video/")
}

func generateRandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
