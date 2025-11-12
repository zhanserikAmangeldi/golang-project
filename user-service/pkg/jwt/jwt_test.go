package jwt

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	secretKey := "supersecretkey"
	manager := NewTokenManager(secretKey)

	userId := int64(123)
	username := "testuser"
	email := "test@example.com"

	tokenStr, expiresAt, err := manager.GenerateAccessToken(userId, username, email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}
	expectedExp := time.Now().Add(15 * time.Minute)
	if expiresAt.Sub(expectedExp) > 2*time.Second {
		t.Errorf("expected expiry around %v, got %v", expectedExp, expiresAt)
	}

	claims, err := manager.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if claims.UserId != userId || claims.Username != username || claims.Email != email {
		t.Errorf("claims mismatch: got %+v", claims)
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	secretKey := "refreshsecret"
	manager := NewTokenManager(secretKey)

	userId := int64(456)
	username := "refreshuser"
	email := "refresh@example.com"

	tokenStr, expiresAt, err := manager.GenerateRefreshToken(userId, username, email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty refresh token")
	}
	expectedExp := time.Now().Add(7 * 24 * time.Hour)
	if expiresAt.Sub(expectedExp) > 2*time.Second {
		t.Errorf("expected expiry around %v, got %v", expectedExp, expiresAt)
	}

	claims, err := manager.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if claims.UserId != userId || claims.Username != username || claims.Email != email {
		t.Errorf("claims mismatch: got %+v", claims)
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	manager := NewTokenManager("secret1")
	managerWrong := NewTokenManager("wrongsecret")

	tokenStr, _, err := manager.GenerateAccessToken(1, "user", "user@example.com")
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	claims, err := managerWrong.ValidateToken(tokenStr)
	if err == nil || err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
	if claims != nil {
		t.Error("expected nil claims for invalid signature")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	secret := "expiringkey"
	manager := NewTokenManager(secret)

	expiredTime := time.Now().Add(-1 * time.Minute)
	claims := Claims{
		UserId:   1,
		Username: "expired",
		Email:    "exp@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiredTime),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	c, err := manager.ValidateToken(tokenStr)
	if err != ErrExpiredToken {
		t.Fatalf("expected ErrExpiredToken, got %v", err)
	}
	if c != nil {
		t.Error("expected nil claims for expired token")
	}
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	manager := NewTokenManager("secretkey")

	c, err := manager.ValidateToken("invalid.token.structure")
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
	if c != nil {
		t.Error("expected nil claims for malformed token")
	}

	c, err = manager.ValidateToken(strings.Repeat("a", 50))
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
	if c != nil {
		t.Error("expected nil claims for invalid token string")
	}
}
