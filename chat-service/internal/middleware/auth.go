package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// AuthMiddleware to check token signatures and check revoke
func AuthMiddleware(jwtSecret string, redisClient *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, tokenString, err := GetToken(r, jwtSecret)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to extract token with error: %s", err.Error()), http.StatusUnauthorized)
				return
			}

			err = CheckRevokedToken(r, redisClient, tokenString)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to validate token with error: %s", err.Error()), http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}

			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				http.Error(w, "invalid user_id in token", http.StatusUnauthorized)
				return
			}
			userID := int64(userIDFloat)

			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			r.Header.Set("X-User-ID", strconv.FormatInt(userID, 10))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts user ID from request context
func GetUserID(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	return userID, ok
}

// GetToken extracts token from request context
func GetToken(r *http.Request, jwtSecret string) (*jwt.Token, string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, "", errors.New("missing authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return nil, "", errors.New("invalid authorization format")
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, "", errors.New("invalid token")
	}

	return token, tokenString, nil
}

func CheckRevokedToken(r *http.Request, redisClient *redis.Client, tokenString string) error {
	exists, err := redisClient.Exists(r.Context(), "revoked:"+tokenString).Result()
	if err != nil {
		return errors.New("redis error")
	}
	if exists > 0 {
		return errors.New("revoked token")
	}

	return nil
}
