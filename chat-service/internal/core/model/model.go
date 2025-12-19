package model

import (
    "time"
)

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
    ID             int64     `json:"id" db:"id"`
    ConversationID int64     `json:"conversation_id" db:"conversation_id"`
    SenderID       int64     `json:"sender_id" db:"sender_id"`
    Content        string    `json:"content" db:"content"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
    MessageType    string    `json:"message_type" db:"message_type"`
    FileURL        *string   `json:"file_url,omitempty" db:"file_url"`
    FileName       *string   `json:"file_name,omitempty" db:"file_name"`
    MimeType       *string   `json:"mime_type,omitempty" db:"mime_type"`
    FileSize       *int64    `json:"file_size,omitempty" db:"file_size"`
}