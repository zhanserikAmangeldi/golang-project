package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhanserikAmangeldi/user-service/internal/models"
	"github.com/zhanserikAmangeldi/user-service/internal/repository"
)

const (
	roleKey = "user_role"
)

func RequireRole(userRepo *repository.UserRepository, requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify user role"})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range requiredRoles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Set(roleKey, user.Role)

		c.Next()
	}
}

func RequireAdmin(userRepo *repository.UserRepository) gin.HandlerFunc {
	return RequireRole(userRepo, models.RoleAdmin)
}

func RequireModerator(userRepo *repository.UserRepository) gin.HandlerFunc {
	return RequireRole(userRepo, models.RoleAdmin, models.RoleModerator)
}

func GetUserRole(c *gin.Context) string {
	role, exists := c.Get(roleKey)
	if !exists {
		return ""
	}
	return role.(string)
}

func CheckBan(banRepo *repository.BanRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		isBanned, ban, err := banRepo.IsUserBanned(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check ban status"})
			c.Abort()
			return
		}

		if isBanned {
			response := gin.H{
				"error":   "account_banned",
				"message": "Your account has been banned",
				"reason":  ban.Reason,
			}

			if !ban.IsPermanent && ban.ExpiresAt != nil {
				response["expires_at"] = ban.ExpiresAt
			} else {
				response["permanent"] = true
			}

			c.JSON(http.StatusForbidden, response)
			c.Abort()
			return
		}

		c.Next()
	}
}
