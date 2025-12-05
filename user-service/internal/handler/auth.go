package handler

import (
	"errors"
	"fmt"
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
// @Summary Register a new user
// @Description Create new account and send verification email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterUserRequest true "User registration data"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/register [post]
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
// @Summary Login user
// @Description Authenticate user and return access + refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/login [post]
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
// @Summary Refresh JWT tokens
// @Description Refresh token pair using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/refresh [post]
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
// @Summary Logout from current session
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.TokensRequest true "Access + refresh tokens"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/logout [post]
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
// @Summary Logout user from all devices
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := middleware.GetUserID(c)
	fmt.Println(userID)
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
// @Summary List active sessions for current user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Param current_token query string false "Current session refresh token"
// @Success 200 {object} dto.SessionListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/sessions [get]
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
	c.JSON(http.StatusOK, dto.SessionListResponse{
       Sessions: sessions.Sessions,
       Total:    sessions.Total,
    })

	// c.JSON(http.StatusOK, sessions)
}
