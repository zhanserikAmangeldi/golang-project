package repository

import (
	"database/sql"
)

type Conversation struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsGroup   bool   `json:"is_group"`
	CreatedAt string `json:"created_at"`
}

type ConversationRepo struct {
	db *sql.DB
}

func NewConversationRepo(db *sql.DB) *ConversationRepo {
	return &ConversationRepo{db: db}
}

func (r *ConversationRepo) Create(name string, isGroup bool) (*Conversation, error) {
	conv := &Conversation{}
	err := r.db.QueryRow(`
        INSERT INTO conversations (name, is_group) VALUES ($1, $2)
        RETURNING id, name, is_group, created_at
    `, name, isGroup).Scan(&conv.ID, &conv.Name, &conv.IsGroup, &conv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return conv, nil
}

func (r *ConversationRepo) AddParticipants(convID string, userIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO participants (conversation_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, uid := range userIDs {
		if _, err := stmt.Exec(convID, uid); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *ConversationRepo) List() ([]Conversation, error) {
	rows, err := r.db.Query(`SELECT id, name, is_group, created_at FROM conversations ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Conversation
	for rows.Next() {
		var c Conversation
		if err := rows.Scan(&c.ID, &c.Name, &c.IsGroup, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *ConversationRepo) GetByID(id string) (*Conversation, error) {
	conv := &Conversation{}
	err := r.db.QueryRow(`SELECT id, name, is_group, created_at FROM conversations WHERE id=$1`, id).
		Scan(&conv.ID, &conv.Name, &conv.IsGroup, &conv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return conv, nil
}
