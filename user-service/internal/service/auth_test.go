package service

import (
	"context"
	"github.com/redis/go-redis/v9"
	jwt "github.com/zhanserikAmangeldi/user-service/pkg/jwt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/zhanserikAmangeldi/user-service/internal/dto"
	"github.com/zhanserikAmangeldi/user-service/internal/models"
	"github.com/zhanserikAmangeldi/user-service/internal/repository"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	user.ID = 1
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastSeen(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetAvatarURL(ctx context.Context, userID int64) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockUserRepository) UpdateAvatar(ctx context.Context, userID int64, objectName string) error {
	args := m.Called(ctx, userID, objectName)
	return args.Error(0)
}

func (m *MockUserRepository) MarkVerified(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *repository.Session) error {
	args := m.Called(ctx, session)
	session.ID = 1
	session.CreatedAt = time.Now()
	return args.Error(0)
}

func (m *MockSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*repository.Session, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Session), args.Error(1)
}

func (m *MockSessionRepository) Revoke(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockSessionRepository) RevokeAllByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) GetAllByUserID(ctx context.Context, userID int64) ([]*repository.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*repository.Session), args.Error(1)
}

func (m *MockSessionRepository) UpdateAccessToken(ctx context.Context, refreshToken, newAccessToken string) error {
	args := m.Called(ctx, refreshToken, newAccessToken)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return int64(args.Int(0)), args.Error(1)
}

type MockEmailVerificationRepository struct {
	mock.Mock
}

func (m *MockEmailVerificationRepository) Create(ctx context.Context, ev *models.EmailVerification) error {
	args := m.Called(ctx, ev)
	ev.ID = 1
	ev.CreatedAt = time.Now()
	return args.Error(0)
}

func (m *MockEmailVerificationRepository) GetByToken(ctx context.Context, token string) (*models.EmailVerification, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EmailVerification), args.Error(1)
}

func (m *MockEmailVerificationRepository) MarkVerified(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendVerificationEmail(to, username, token string) error {
	args := m.Called(to, username, token)
	return args.Error(0)
}

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) Exists(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)

	cmd := redis.NewIntResult(int64(args.Int(0)), args.Error(1))
	return cmd
}

func TestRegister_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockEmailRepo := new(MockEmailVerificationRepository)
	mockEmailSender := new(MockEmailSender)
	mockRedis := new(MockRedisClient)

	tokenManager := &mockTokenManager{}

	service := &AuthService{
		userRepo:     mockUserRepo,
		sessionRepo:  mockSessionRepo,
		emailRepo:    mockEmailRepo,
		emailSender:  mockEmailSender,
		redisClient:  mockRedis,
		tokenManager: tokenManager,
	}

	ctx := context.Background()
	req := &dto.RegisterUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
		Return(nil)

	mockEmailRepo.On("Create", ctx, mock.AnythingOfType("*models.EmailVerification")).
		Return(nil)

	mockEmailSender.On("SendVerificationEmail", req.Email, req.Username, mock.AnythingOfType("string")).
		Return(nil)

	mockSessionRepo.On("Create", ctx, mock.AnythingOfType("*repository.Session")).
		Return(nil)

	authResp, err := service.Register(ctx, req, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, authResp)
	assert.NotEmpty(t, authResp.AccessToken)
	assert.NotEmpty(t, authResp.RefreshToken)
	assert.Equal(t, "testuser", authResp.User.Username)

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
	mockEmailRepo.AssertExpectations(t)
	mockEmailSender.AssertExpectations(t)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockEmailRepo := new(MockEmailVerificationRepository)
	mockEmailSender := new(MockEmailSender)
	mockRedis := new(MockRedisClient)
	tokenManager := &mockTokenManager{}

	service := &AuthService{
		userRepo:     mockUserRepo,
		sessionRepo:  mockSessionRepo,
		emailRepo:    mockEmailRepo,
		emailSender:  mockEmailSender,
		redisClient:  mockRedis,
		tokenManager: tokenManager,
	}

	ctx := context.Background()
	req := &dto.RegisterUserRequest{
		Username: "existinguser",
		Email:    "existing@example.com",
		Password: "password123",
	}

	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
		Return(repository.ErrUserAlreadyExists)

	authResp, err := service.Register(ctx, req, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Equal(t, ErrAlreadyUserExists, err)

	mockUserRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockRedis := new(MockRedisClient)
	tokenManager := &mockTokenManager{}

	service := &AuthService{
		userRepo:     mockUserRepo,
		sessionRepo:  mockSessionRepo,
		redisClient:  mockRedis,
		tokenManager: tokenManager,
	}

	ctx := context.Background()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	existingUser := &models.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	req := &dto.LoginRequest{
		Login:    "test@example.com",
		Password: password,
	}

	mockUserRepo.On("GetByEmail", ctx, "test@example.com").
		Return(existingUser, nil)

	mockSessionRepo.On("Create", ctx, mock.AnythingOfType("*repository.Session")).
		Return(nil)

	mockUserRepo.On("UpdateLastSeen", ctx, int64(1)).
		Return(nil)

	authResp, err := service.Login(ctx, req, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, authResp)
	assert.NotEmpty(t, authResp.AccessToken)
	assert.NotEmpty(t, authResp.RefreshToken)
	assert.Equal(t, int64(1), authResp.User.ID)

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockRedis := new(MockRedisClient)
	tokenManager := &mockTokenManager{}

	service := &AuthService{
		userRepo:     mockUserRepo,
		sessionRepo:  mockSessionRepo,
		redisClient:  mockRedis,
		tokenManager: tokenManager,
	}

	ctx := context.Background()
	req := &dto.LoginRequest{
		Login:    "test@example.com",
		Password: "wrongpassword",
	}

	mockUserRepo.On("GetByEmail", ctx, "test@example.com").
		Return(nil, repository.ErrUserNotFound)

	authResp, err := service.Login(ctx, req, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Equal(t, ErrInvalidCredentials, err)

	mockUserRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockRedis := new(MockRedisClient)
	tokenManager := &mockTokenManager{}

	service := &AuthService{
		userRepo:     mockUserRepo,
		sessionRepo:  mockSessionRepo,
		redisClient:  mockRedis,
		tokenManager: tokenManager,
	}

	ctx := context.Background()
	correctPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	existingUser := &models.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	req := &dto.LoginRequest{
		Login:    "test@example.com",
		Password: "wrongpassword",
	}

	mockUserRepo.On("GetByEmail", ctx, "test@example.com").
		Return(existingUser, nil)

	authResp, err := service.Login(ctx, req, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Equal(t, ErrInvalidCredentials, err)

	mockUserRepo.AssertExpectations(t)
}

type mockTokenManager struct{}

func (m *mockTokenManager) GenerateAccessToken(userID int64, username, email string) (string, time.Time, error) {
	return "mock-access-token", time.Now().Add(15 * time.Minute), nil
}

func (m *mockTokenManager) GenerateRefreshToken(userID int64, username, email string) (string, time.Time, error) {
	return "mock-refresh-token", time.Now().Add(7 * 24 * time.Hour), nil
}

func (m *mockTokenManager) ValidateToken(token string) (*jwt.Claims, error) {
	return &jwt.Claims{
		UserId:   1,
		Username: "testuser",
		Email:    "test@example.com",
	}, nil
}
