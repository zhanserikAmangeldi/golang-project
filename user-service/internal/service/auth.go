package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhanserikAmangeldi/user-service/internal/dto"
	"github.com/zhanserikAmangeldi/user-service/internal/models"
	"github.com/zhanserikAmangeldi/user-service/internal/repository"
	"github.com/zhanserikAmangeldi/user-service/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAlreadyUserExists  = errors.New("user already exists")
)

type AuthService struct {
	userRepo     *repository.UserRepository
	sessionRepo  *repository.SessionRepository
	tokenManager *jwt.TokenManager
}

func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	tokenManager *jwt.TokenManager,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		tokenManager: tokenManager,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterUserRequest, userAgent, ipAddress *string) (*dto.AuthResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if req.DisplayName != "" {
		user.DisplayName = &req.DisplayName
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrAlreadyUserExists
		}
		return nil, err
	}

	accessToken, expiresAt, err := s.tokenManager.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.tokenManager.GenerateRefreshToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	session := &repository.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    expiresAt,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(expiresAt.Sub(expiresAt.Add(-24 * 3600)).Seconds()),
		User:         user,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest, userAgent, ipAddress *string) (*dto.AuthResponse, error) {
	var user *models.User
	var err error

	if strings.Contains(req.Login, "@") {
		user, err = s.userRepo.GetByEmail(ctx, req.Login)
	} else {
		user, err = s.userRepo.GetByUsername(ctx, req.Login)
	}

	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, expiresAt, err := s.tokenManager.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpiresAt, err := s.tokenManager.GenerateRefreshToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	session := &repository.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    refreshExpiresAt,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	_ = s.userRepo.UpdateLastSeen(ctx, user.ID)

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(expiresAt.Sub(expiresAt.Add(-24 * 3600)).Seconds()),
		User:         user,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string, userAgent, ipAddress *string) (*dto.AuthResponse, error) {
	_, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil, errors.New("invalid refresh token")
		}
		if errors.Is(err, repository.ErrSessionExpired) {
			return nil, errors.New("refresh token expired")
		}
		if errors.Is(err, repository.ErrSessionRevoked) {
			return nil, errors.New("session revoked")
		}
		return nil, err
	}

	claims, err := s.tokenManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserId)
	if err != nil {
		return nil, err
	}

	newAccessToken, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	newRefreshToken, refreshExpiresAt, err := s.tokenManager.GenerateRefreshToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	if err := s.sessionRepo.Revoke(ctx, refreshToken); err != nil {
		return nil, err
	}

	newSession := &repository.Session{
		UserID:       user.ID,
		RefreshToken: newRefreshToken,
		AccessToken:  newAccessToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    refreshExpiresAt,
	}

	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(accessExpiresAt.Sub(time.Now()).Seconds()),
		User:         user,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.sessionRepo.Revoke(ctx, refreshToken)
}

func (s *AuthService) LogoutAll(ctx context.Context, userID int64) error {
	return s.sessionRepo.RevokeAllByUserID(ctx, userID)
}

func (s *AuthService) GetActiveSessions(ctx context.Context, userID int64, currentRefreshToken string) (*models.SessionListResponse, error) {
	sessions, err := s.sessionRepo.GetAllByUserID(ctx, userID)
	fmt.Println("check 1")
	if err != nil {
		return nil, err
	}
	fmt.Print("check 2")

	sessionInfos := make([]*models.SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		sessionInfos = append(sessionInfos, &models.SessionInfo{
			ID:        sess.ID,
			UserAgent: sess.UserAgent,
			IPAddress: sess.IPAddress,
			CreatedAt: sess.CreatedAt,
			ExpiresAt: sess.ExpiresAt,
			IsCurrent: sess.RefreshToken == currentRefreshToken,
		})
	}

	return &models.SessionListResponse{
		Sessions: sessionInfos,
		Total:    len(sessionInfos),
	}, nil
}
