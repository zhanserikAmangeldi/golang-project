package dto

import (
	"github.com/zhanserikAmangeldi/user-service/internal/models"
)

type RegisterUserRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=32"`
	DisplayName string `json:"display_name,omitempty" binding:"max=50"`
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"` // может быть email или username
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         *models.User `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateUserRequest struct {
	DisplayName *string `json:"display_name,omitempty" binding:"omitempty,max=100"`
	AvatarURL   *string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	Bio         *string `json:"bio,omitempty" binding:"omitempty,max=500"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=online offline away busy"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
