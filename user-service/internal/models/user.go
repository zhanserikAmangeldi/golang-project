package models

import "time"

type User struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	DisplayName  *string    `json:"display_name,omitempty"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	Bio          *string    `json:"bio,omitempty"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type UserBan struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	BannedBy    int64      `json:"banned_by"`
	Reason      string     `json:"reason"`
	BannedAt    time.Time  `json:"banned_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsPermanent bool       `json:"is_permanent"`
	UnbannedAt  *time.Time `json:"unbanned_at,omitempty"`
	UnbannedBy  *int64     `json:"unbanned_by,omitempty"`
}

type AdminAuditLog struct {
	ID         int64       `json:"id"`
	AdminID    int64       `json:"admin_id"`
	Action     string      `json:"action"`
	TargetType string      `json:"target_type"`
	TargetID   *int64      `json:"target_id,omitempty"`
	Details    interface{} `json:"details,omitempty"`
	IPAddress  *string     `json:"ip_address,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
}

const (
	RoleUser      = "user"
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
)

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsModerator() bool {
	return u.Role == RoleModerator || u.Role == RoleAdmin
}

func (b *UserBan) IsActive() bool {
	if b.UnbannedAt != nil {
		return false
	}
	if b.IsPermanent {
		return true
	}
	if b.ExpiresAt != nil && time.Now().Before(*b.ExpiresAt) {
		return true
	}
	return false
}
