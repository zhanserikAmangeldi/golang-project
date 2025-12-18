package dto

type BanUserRequest struct {
	UserID          int64  `json:"user_id" binding:"required"`
	Reason          string `json:"reason" binding:"required,min=10,max=500"`
	DurationMinutes int    `json:"duration_minutes"` // 0 = permanent
}

type UpdateRoleRequest struct {
	UserID int64  `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=user admin moderator"`
}

type UserListResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type UserResponse struct {
	ID          int64   `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	Status      string  `json:"status"`
	DisplayName *string `json:"display_name,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type BanInfo struct {
	IsBanned    bool    `json:"is_banned"`
	Reason      string  `json:"reason,omitempty"`
	BannedAt    string  `json:"banned_at,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
	IsPermanent bool    `json:"is_permanent"`
}
