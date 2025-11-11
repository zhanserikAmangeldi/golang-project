package repository

import (
	"database/sql"
	"time"
)

type Message struct {
	ID             string     `json:"id"`
	ConversationID string     `json:"conversation_id"`
	SenderID       string     `json:"sender_id"`
	Content        string     `json:"content"`
	MessageType    string     `json:"message_type"`
	CreatedAt      time.Time  `json:"created_at"`
	EditedAt       *time.Time `json:"edited_at"`
	IsDeleted      bool       `json:"is_deleted"`
}

type MessageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Save(conversationID, senderID, content, msgType string) (*Message, error) {
	m := &Message{}
	err := r.db.QueryRow(`
        INSERT INTO messages (conversation_id, sender_id, content, message_type)
        VALUES ($1, $2, $3, $4)
        RETURNING id, conversation_id, sender_id, content, message_type, created_at
    `, conversationID, senderID, content, msgType).Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.MessageType, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MessageRepo) ListByConversation(conversationID string, limit int, before time.Time) ([]Message, error) {
	var rows *sql.Rows
	var err error
	if !before.IsZero() {
		rows, err = r.db.Query(`
            SELECT id, conversation_id, sender_id, content, message_type, created_at, edited_at, is_deleted
            FROM messages
            WHERE conversation_id = $1 AND created_at < $2
            ORDER BY created_at DESC
            LIMIT $3`, conversationID, before, limit)
	} else {
		rows, err = r.db.Query(`
            SELECT id, conversation_id, sender_id, content, message_type, created_at, edited_at, is_deleted
            FROM messages
            WHERE conversation_id = $1
            ORDER BY created_at DESC
            LIMIT $2`, conversationID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var m Message
		var edited sql.NullTime
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.MessageType, &m.CreatedAt, &edited, &m.IsDeleted); err != nil {
			return nil, err
		}
		if edited.Valid {
			t := edited.Time
			m.EditedAt = &t
		}
		out = append(out, m)
	}
	return out, nil
}
