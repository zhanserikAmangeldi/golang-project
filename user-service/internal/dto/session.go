package dto

import "github.com/zhanserikAmangeldi/user-service/internal/models"

// SessionListResponse — то, что Swagger будет показывать при /auth/sessions
type SessionListResponse struct {
	Sessions []*models.SessionInfo `json:"sessions"`
	Total    int                   `json:"total"`
}
