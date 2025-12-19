package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zhanserikAmangeldi/user-service/internal/service"
	"net/http"
)

type EmailVerificationHandler struct {
	authService *service.AuthService
}

func NewEmailVerificationHandler(authService *service.AuthService) *EmailVerificationHandler {
	return &EmailVerificationHandler{authService: authService}
}

// VerifyEmail godoc
// @Summary Подтверждение Email
// @Description Проверяет токен из ссылки, отправленной на почту, и активирует аккаунт
// @Tags auth
// @Produce json
// @Param token query string true "Токен подтверждения"
// @Success 200 {object} map[string]string "message: email verified successfully"
// @Failure 400 {object} map[string]string "error: invalid or expired token"
// @Router /verify-email [get]
func (h *EmailVerificationHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no token provided"})
		return
	}

	err := h.authService.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}