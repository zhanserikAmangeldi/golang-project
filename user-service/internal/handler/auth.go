package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/zhanserikAmangeldi/user-service/internal/dto"
	"github.com/zhanserikAmangeldi/user-service/internal/middleware"
	"github.com/zhanserikAmangeldi/user-service/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func getClientInfo(c *gin.Context) (*string, *string) {
	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	var userAgentStr *string
	var ipPtr *string

	if userAgent != "" {
		userAgentStr = &userAgent
	}
	if ip != "" {
		ipPtr = &ip
	}

	return userAgentStr, ipPtr
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterUserRequest true "Данные"
// @Success 201 {object} dto.AuthResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	userAgent, ip := getClientInfo(c)
	authResp, err := h.authService.Register(c.Request.Context(), &req, userAgent, ip)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyUserExists) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "user_exists",
				Message: "User with this email or username already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_server",
			Message: "Failed to register user",
		})
		return
	}

	c.JSON(http.StatusCreated, authResp)
}

// Login godoc
// @Summary Вход пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Логин"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	userAgent, ip := getClientInfo(c)
	authResp, err := h.authService.Login(c.Request.Context(), &req, userAgent, ip)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Invalid email/username or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to login",
		})
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// RefreshToken godoc
// @Summary Обновить токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Токен"
// @Success 200 {object} dto.AuthResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	userAgent, ip := getClientInfo(c)
	authResp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken, userAgent, ip)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "invalid_token",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// Logout godoc
// @Summary Выход
// @Tags auth
// @Security ApiKeyAuth
// @Accept json
// @Param request body dto.TokensRequest true "Токены"
// @Success 200 {string} string "ok"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.TokensRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	err := h.authService.Logout(c.Request.Context(), req.RefreshToken, req.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_server",
			Message: "Failed to logout",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// LogoutAll godoc
// @Summary Выход со всех устройств
// @Tags auth
// @Security ApiKeyAuth
// @Success 200 {string} string "ok"
// @Router /api/v1/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "unauthorized",
		})
		return
	}

	err := h.authService.LogoutAll(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to logout from all devices",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out from all devices successfully",
	})
}

// GetActiveSessions godoc
// @Summary Активные сессии
// @Tags auth
// @Security ApiKeyAuth
// @Param current_token query string false "Token"
// @Success 200 {array} interface{}
// @Router /api/v1/auth/sessions [get]
func (h *AuthHandler) GetActiveSessions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "unauthorized",
		})
		return
	}

	currentRefreshToken := c.Query("current_token")

	sessions, err := h.authService.GetActiveSessions(c.Request.Context(), userID, currentRefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "internal_error",
		})
		return
	}

	c.JSON(http.StatusOK, sessions)
}