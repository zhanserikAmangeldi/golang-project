package repository

import (
	"context"
	"database/sql"
	"fmt"

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
	query := `INSERT INTO conversations (is_group, name, created_at) VALUES ($1, $2, $3) RETURNING id`
	return r.db.QueryRowContext(ctx, query, conv.IsGroup, conv.Name, conv.CreatedAt).Scan(&conv.ID)
}

func (r *PostgresRepository) AddParticipant(ctx context.Context, part *model.Participant) error {
	query := `INSERT INTO participants (conversation_id, user_id, joined_at) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, part.ConversationID, part.UserID, part.JoinedAt)
	return err
}

func (r *PostgresRepository) SaveMessage(ctx context.Context, msg *model.Message) error {
	query := `
		INSERT INTO messages (conversation_id, sender_id, content, message_type, file_url, file_name, file_size, mime_type, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, query,
		msg.ConversationID,
		msg.SenderID,
		msg.Content,
		msg.MessageType,
		msg.FileURL,
		msg.FileName,
		msg.FileSize,
		msg.MimeType,
		msg.CreatedAt,
	).Scan(&msg.ID)
}

func (r *PostgresRepository) GetMessages(ctx context.Context, conversationID int64, limit, offset int) ([]model.Message, error) {
	var messages []model.Message
	query := `
		SELECT id, conversation_id, sender_id, content, message_type, 
		       file_url, file_name, file_size, mime_type, 
		       created_at, edited_at, deleted_at
		FROM messages 
		WHERE conversation_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &messages, query, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}

	for i := range messages {
		messages[i].ReadBy, _ = r.GetMessageReads(ctx, messages[i].ID)
		messages[i].Reactions, _ = r.GetMessageReactions(ctx, messages[i].ID)
	}

	return messages, nil
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

func (r *PostgresRepository) GetUserConversations(ctx context.Context, userID int64, limit, offset int) ([]model.ConversationWithLastMessage, error) {
	query := `
		SELECT DISTINCT
			c.id,
			c.is_group,
			c.name,
			c.created_at,
			(
				SELECT COUNT(*) 
				FROM messages m
				LEFT JOIN message_reads mr ON m.id = mr.message_id AND mr.user_id = $1
				WHERE m.conversation_id = c.id 
				AND m.sender_id != $1
				AND m.deleted_at IS NULL
				AND mr.message_id IS NULL
			) as unread_count
		FROM conversations c
		JOIN participants p ON c.id = p.conversation_id
		WHERE p.user_id = $1
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []model.ConversationWithLastMessage
	for rows.Next() {
		var conv model.ConversationWithLastMessage
		err := rows.Scan(&conv.ID, &conv.IsGroup, &conv.Name, &conv.CreatedAt, &conv.UnreadCount)
		if err != nil {
			return nil, err
		}

		lastMsg, _ := r.GetLastMessage(ctx, conv.ID)
		conv.LastMessage = lastMsg

		conv.ParticipantIDs, _ = r.GetParticipants(ctx, conv.ID)

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (r *PostgresRepository) GetLastMessage(ctx context.Context, conversationID int64) (*model.Message, error) {
	var msg model.Message
	query := `
		SELECT id, conversation_id, sender_id, content, message_type,
		       file_url, file_name, file_size, mime_type,
		       created_at, edited_at, deleted_at
		FROM messages
		WHERE conversation_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &msg, query, conversationID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *PostgresRepository) MarkMessageAsRead(ctx context.Context, messageID, userID int64) error {
	query := `
		INSERT INTO message_reads (message_id, user_id, read_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (message_id, user_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, messageID, userID)
	return err
}

func (r *PostgresRepository) GetMessageReads(ctx context.Context, messageID int64) ([]int64, error) {
	var userIDs []int64
	query := `SELECT user_id FROM message_reads WHERE message_id = $1`
	err := r.db.SelectContext(ctx, &userIDs, query, messageID)
	return userIDs, err
}

func (r *PostgresRepository) AddReaction(ctx context.Context, messageID, userID int64, reaction string) error {
	query := `
		INSERT INTO message_reactions (message_id, user_id, reaction, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (message_id, user_id, reaction) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, messageID, userID, reaction)
	return err
}

func (r *PostgresRepository) RemoveReaction(ctx context.Context, messageID, userID int64, reaction string) error {
	query := `DELETE FROM message_reactions WHERE message_id = $1 AND user_id = $2 AND reaction = $3`
	_, err := r.db.ExecContext(ctx, query, messageID, userID, reaction)
	return err
}

func (r *PostgresRepository) GetMessageReactions(ctx context.Context, messageID int64) ([]model.Reaction, error) {
	var reactions []model.Reaction
	query := `SELECT id, message_id, user_id, reaction, created_at FROM message_reactions WHERE message_id = $1`
	err := r.db.SelectContext(ctx, &reactions, query, messageID)
	return reactions, err
}

func (r *PostgresRepository) EditMessage(ctx context.Context, messageID int64, newContent string) error {
	query := `
		UPDATE messages 
		SET content = $2, edited_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, messageID, newContent)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("message not found or already deleted")
	}
	return nil
}

func (r *PostgresRepository) DeleteMessage(ctx context.Context, messageID int64) error {
	query := `UPDATE messages SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("message not found or already deleted")
	}
	return nil
}

func (r *PostgresRepository) GetMessageByID(ctx context.Context, messageID int64) (*model.Message, error) {
	var msg model.Message
	query := `
		SELECT id, conversation_id, sender_id, content, message_type,
		       file_url, file_name, file_size, mime_type,
		       created_at, edited_at, deleted_at
		FROM messages
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, &msg, query, messageID)
	if err != nil {
		return nil, err
	}

	msg.ReadBy, _ = r.GetMessageReads(ctx, msg.ID)
	msg.Reactions, _ = r.GetMessageReactions(ctx, msg.ID)

	return &msg, nil
}
