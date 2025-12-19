package handler

import (
	"fmt"
	"github.com/zhanserikAmangeldi/user-service/internal/middleware"
	"github.com/zhanserikAmangeldi/user-service/internal/repository"
	"github.com/zhanserikAmangeldi/user-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

type MinioHandler struct {
	MinioService *service.Minio
	UserRepo     *repository.UserRepository
}

func NewMinioHandler(minioService *service.Minio, userRepo *repository.UserRepository) *MinioHandler {
	return &MinioHandler{
		MinioService: minioService,
		UserRepo:     userRepo,
	}
}

// UploadAvatar godoc
// @Summary Загрузить аватар
// @Security BearerAuth
// @Tags users
// @Accept multipart/form-data
// @Param avatar formData file true "Файл изображения"
// @Success 200 {object} map[string]string
// @Router /api/v1/users/upload-avatar [post]
func (m *MinioHandler) UploadAvatar(c *gin.Context) {
	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open file"})
		return
	}
	defer file.Close()

	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	objectName := fmt.Sprintf("%v/%s", userID, "avatar")
	contentType := fileHeader.Header.Get("Content-Type")

	_, err = m.MinioService.MinioClient.PutObject(
		c.Request.Context(),
		"avatars",
		objectName,
		file,
		fileHeader.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = m.UserRepo.UpdateAvatar(c.Request.Context(), userID, objectName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user avatar URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Avatar uploaded successfully", "path": objectName})
}

// GetAvatar godoc
// @Summary Получить файл аватара
// @Security BearerAuth
// @Tags users
// @Produce image/png
// @Success 200 {string} binary
// @Router /api/v1/users/get-avatar [get]
func (m *MinioHandler) GetAvatar(c *gin.Context) {
	userID := middleware.GetUserID(c)

	url, err := m.UserRepo.GetAvatarURL(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get avatar URL"})
		return
	}

	object, err := m.MinioService.MinioClient.GetObject(
		c.Request.Context(),
		"avatars",
		url,
		minio.GetObjectOptions{},
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	defer object.Close()

	info, err := object.Stat()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found or unreadable"})
		return
	}

	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf("inline; filename=avatar"),
	}

	c.DataFromReader(
		http.StatusOK,
		info.Size,
		info.ContentType,
		object,
		extraHeaders,
	)
}