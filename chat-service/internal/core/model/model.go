package model

import "time"

type Conversation struct {
	ID        int64     `json:"id" db:"id"`
	IsGroup   bool      `json:"is_group" db:"is_group"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Participant struct {
	ConversationID int64     `json:"conversation_id" db:"conversation_id"`
	UserID         int64     `json:"user_id" db:"user_id"`
	JoinedAt       time.Time `json:"joined_at" db:"joined_at"`
}

type Message struct {
	ID             int64      `json:"id" db:"id"`
	ConversationID int64      `json:"conversation_id" db:"conversation_id"`
	SenderID       int64      `json:"sender_id" db:"sender_id"`
	Content        string     `json:"content" db:"content"`
	MessageType    string     `json:"message_type" db:"message_type"`
	FileURL        *string    `json:"file_url,omitempty" db:"file_url"`
	FileName       *string    `json:"file_name,omitempty" db:"file_name"`
	FileSize       *int64     `json:"file_size,omitempty" db:"file_size"`
	MimeType       *string    `json:"mime_type,omitempty" db:"mime_type"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	EditedAt       *time.Time `json:"edited_at,omitempty" db:"edited_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	ReadBy         []int64    `json:"read_by,omitempty" db:"-"`
	Reactions      []Reaction `json:"reactions,omitempty" db:"-"`
}

type MessageRead struct {
	MessageID int64     `json:"message_id" db:"message_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	ReadAt    time.Time `json:"read_at" db:"read_at"`
}

type Reaction struct {
	ID        int64     `json:"id" db:"id"`
	MessageID int64     `json:"message_id" db:"message_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Reaction  string    `json:"reaction" db:"reaction"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ConversationWithLastMessage struct {
	ID             int64     `json:"id" db:"id"`
	IsGroup        bool      `json:"is_group" db:"is_group"`
	Name           string    `json:"name" db:"name"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	LastMessage    *Message  `json:"last_message,omitempty"`
	UnreadCount    int       `json:"unread_count"`
	ParticipantIDs []int64   `json:"participant_ids,omitempty"`
}

type TypingEvent struct {
	ConversationID int64  `json:"conversation_id"`
	UserID         int64  `json:"user_id"`
	Username       string `json:"username"`
	IsTyping       bool   `json:"is_typing"`
}

type OnlineStatusEvent struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Status   string `json:"status"` // online, offline, away
}

type WSMessage struct {
	Type    string      `json:"type"` // message, typing, status, reaction, read_receipt
	Payload interface{} `json:"payload"`
}
