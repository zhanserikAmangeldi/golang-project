package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhanserikAmangeldi/user-service/internal/models"
)

var (
	ErrUserAlreadyBanned = errors.New("user is already banned")
	ErrBanNotFound       = errors.New("ban not found")
)

type BanRepository struct {
	db *pgxpool.Pool
}

func NewBanRepository(db *pgxpool.Pool) *BanRepository {
	return &BanRepository{db: db}
}

func (r *BanRepository) BanUser(ctx context.Context, ban *models.UserBan) error {
	activeBan, err := r.GetActiveBan(ctx, ban.UserID)
	if err != nil && err != ErrBanNotFound {
		return err
	}
	if activeBan != nil {
		return ErrUserAlreadyBanned
	}

	query := `
		INSERT INTO user_bans (user_id, banned_by, reason, expires_at, is_permanent)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, banned_at
	`

	err = r.db.QueryRow(ctx, query,
		ban.UserID,
		ban.BannedBy,
		ban.Reason,
		ban.ExpiresAt,
		ban.IsPermanent,
	).Scan(&ban.ID, &ban.BannedAt)

	return err
}

func (r *BanRepository) GetActiveBan(ctx context.Context, userID int64) (*models.UserBan, error) {
	query := `
		SELECT id, user_id, banned_by, reason, banned_at, expires_at, 
		       is_permanent, unbanned_at, unbanned_by
		FROM user_bans
		WHERE user_id = $1 
		  AND unbanned_at IS NULL
		  AND (is_permanent = TRUE OR expires_at > NOW())
		ORDER BY banned_at DESC
		LIMIT 1
	`

	ban := &models.UserBan{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&ban.ID,
		&ban.UserID,
		&ban.BannedBy,
		&ban.Reason,
		&ban.BannedAt,
		&ban.ExpiresAt,
		&ban.IsPermanent,
		&ban.UnbannedAt,
		&ban.UnbannedBy,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBanNotFound
	}
	if err != nil {
		return nil, err
	}

	return ban, nil
}

func (r *BanRepository) UnbanUser(ctx context.Context, userID, adminID int64) error {
	query := `
		UPDATE user_bans
		SET unbanned_at = NOW(), unbanned_by = $2
		WHERE user_id = $1 AND unbanned_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, userID, adminID)
	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return ErrBanNotFound
	}

	return nil
}

func (r *BanRepository) GetUserBanHistory(ctx context.Context, userID int64) ([]*models.UserBan, error) {
	query := `
		SELECT id, user_id, banned_by, reason, banned_at, expires_at,
		       is_permanent, unbanned_at, unbanned_by
		FROM user_bans
		WHERE user_id = $1
		ORDER BY banned_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bans []*models.UserBan
	for rows.Next() {
		ban := &models.UserBan{}
		err := rows.Scan(
			&ban.ID,
			&ban.UserID,
			&ban.BannedBy,
			&ban.Reason,
			&ban.BannedAt,
			&ban.ExpiresAt,
			&ban.IsPermanent,
			&ban.UnbannedAt,
			&ban.UnbannedBy,
		)
		if err != nil {
			return nil, err
		}
		bans = append(bans, ban)
	}

	return bans, nil
}

func (r *BanRepository) IsUserBanned(ctx context.Context, userID int64) (bool, *models.UserBan, error) {
	ban, err := r.GetActiveBan(ctx, userID)
	if errors.Is(err, ErrBanNotFound) {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}

	return ban.IsActive(), ban, nil
}

func (r *BanRepository) CleanupExpiredBans(ctx context.Context) (int64, error) {
	query := `
		UPDATE user_bans
		SET unbanned_at = NOW()
		WHERE expires_at <= NOW() 
		  AND unbanned_at IS NULL
		  AND is_permanent = FALSE
	`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
