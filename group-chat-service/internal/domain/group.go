package domain

import "time"

type Group struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupMember struct {
	GroupID  int       `json:"group_id" gorm:"primaryKey"`
	UserID   int       `json:"user_id" gorm:"primaryKey"`
	JoinedAt time.Time `json:"joined_at"`
}

type GroupMessage struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	GroupID   int       `json:"group_id"`
	SenderID  int       `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
