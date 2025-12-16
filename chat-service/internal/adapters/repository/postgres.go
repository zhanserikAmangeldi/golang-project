package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/model"
	"github.com/zhanserikAmangeldi/chat-service/internal/core/ports"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) ports.ChatRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateConversation(ctx context.Context, conv *model.Conversation) error {
	query := `INSERT INTO conversations (is_group, created_at) VALUES ($1, $2) RETURNING id`
	return r.db.QueryRowContext(ctx, query, conv.IsGroup, conv.CreatedAt).Scan(&conv.ID)
}

func (r *PostgresRepository) AddParticipant(ctx context.Context, part *model.Participant) error {
	query := `INSERT INTO participants (conversation_id, user_id, joined_at) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, part.ConversationID, part.UserID, part.JoinedAt)
	return err
}

func (r *PostgresRepository) SaveMessage(ctx context.Context, msg *model.Message) error {
	query := `INSERT INTO messages (conversation_id, sender_id, content, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.db.QueryRowContext(ctx, query, msg.ConversationID, msg.SenderID, msg.Content, msg.CreatedAt).Scan(&msg.ID)
}

func (r *PostgresRepository) GetMessages(ctx context.Context, conversationID int64, limit, offset int) ([]model.Message, error) {
	var messages []model.Message
	query := `SELECT * FROM messages WHERE conversation_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &messages, query, conversationID, limit, offset)
	return messages, err
}

func (r *PostgresRepository) GetConversationByID(ctx context.Context, id int64) (*model.Conversation, error) {
	var conv model.Conversation
	query := `SELECT * FROM conversations WHERE id = $1`
	err := r.db.GetContext(ctx, &conv, query, id)
	return &conv, err
}

func (r *PostgresRepository) FindOneToOneConversation(ctx context.Context, user1, user2 int64) (*model.Conversation, error) {
	var conv model.Conversation
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

// GetParticipants fetches all user IDs belonging to a specific conversation
func (r *PostgresRepository) GetParticipants(ctx context.Context, conversationID int64) ([]int64, error) {
	var userIDs []int64
	query := `SELECT user_id FROM participants WHERE conversation_id = $1`
	err := r.db.SelectContext(ctx, &userIDs, query, conversationID)

	return userIDs, err
}

func (r *PostgresRepository) IsParticipant(ctx context.Context, convID, userID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM participants WHERE conversation_id = $1 AND user_id = $2)`
	err := r.db.QueryRowContext(ctx, query, convID, userID).Scan(&exists)
	return exists, err
}
