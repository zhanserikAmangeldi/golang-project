package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"chat-service/internal/core/domain"
	"chat-service/internal/core/ports"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) ports.ChatRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateConversation(ctx context.Context, conv *domain.Conversation) error {
	query := `INSERT INTO conversations (is_group, created_at) VALUES ($1, $2) RETURNING id`
	return r.db.QueryRowContext(ctx, query, conv.IsGroup, conv.CreatedAt).Scan(&conv.ID)
}

func (r *PostgresRepository) AddParticipant(ctx context.Context, part *domain.Participant) error {
	query := `INSERT INTO participants (conversation_id, user_id, joined_at) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, part.ConversationID, part.UserID, part.JoinedAt)
	return err
}

func (r *PostgresRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	query := `INSERT INTO messages (conversation_id, sender_id, content, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.db.QueryRowContext(ctx, query, msg.ConversationID, msg.SenderID, msg.Content, msg.CreatedAt).Scan(&msg.ID)
}

func (r *PostgresRepository) GetMessages(ctx context.Context, conversationID int64, limit, offset int) ([]domain.Message, error) {
	var messages []domain.Message
	query := `SELECT * FROM messages WHERE conversation_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &messages, query, conversationID, limit, offset)
	return messages, err
}

func (r *PostgresRepository) GetConversationByID(ctx context.Context, id int64) (*domain.Conversation, error) {
	var conv domain.Conversation
	query := `SELECT * FROM conversations WHERE id = $1`
	err := r.db.GetContext(ctx, &conv, query, id)
	return &conv, err
}

func (r *PostgresRepository) FindOneToOneConversation(ctx context.Context, user1, user2 int64) (*domain.Conversation, error) {
	var conv domain.Conversation
	query := `
		SELECT c.* 
		FROM conversations c
		JOIN participants p1 ON c.id = p1.conversation_id
		JOIN participants p2 ON c.id = p2.conversation_id
		WHERE c.is_group = false 
		AND p1.user_id = $1 
		AND p2.user_id = $2
	`
	err := r.db.GetContext(ctx, &conv, query, user1, user2)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &conv, err
}
