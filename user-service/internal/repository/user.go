package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhanserikAmangeldi/user-service/internal/models"
	"time"
)

type IUserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateLastSeen(ctx context.Context, userID int64) error
	GetAvatarURL(ctx context.Context, userID int64) (string, error)
	UpdateAvatar(ctx context.Context, userID int64, objectName string) error
	MarkVerified(ctx context.Context, userID int64) error
}

var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, display_name, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.DisplayName,
		"offline",
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_username_key\" (SQLSTATE 23505)" {
			return ErrUserAlreadyExists
		}
		return err
	}

	user.Status = "offline"
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, 
		       bio, role, status, last_seen_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Bio,
		&user.Role,
		&user.Status,
		&user.LastSeenAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url,
		       bio, role, status, last_seen_at, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Bio,
		&user.Role,
		&user.Status,
		&user.LastSeenAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url,
		       bio, role, status, last_seen_at, created_at, updated_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Bio,
		&user.Role,
		&user.Status,
		&user.LastSeenAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetAvatarURL(ctx context.Context, userID int64) (string, error) {
	query := `
		SELECT avatar_url
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var avatarURL string
	err := r.db.QueryRow(ctx, query, userID).Scan(&avatarURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	return avatarURL, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET display_name = $2, avatar_url = $3, bio = $4, status = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query,
		user.ID,
		user.DisplayName,
		user.AvatarURL,
		user.Bio,
		user.Status,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (r *UserRepository) UpdateAvatar(ctx context.Context, userID int64, objectName string) error {
	query := `
		UPDATE users
		SET avatar_url = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at
	`
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	err = r.db.QueryRow(ctx, query,
		userID,
		objectName,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (r *UserRepository) UpdateLastSeen(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET last_seen_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, userID, time.Now())
	return err
}

func (r *UserRepository) MarkVerified(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET is_verified = TRUE, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID int64, newRole string) error {
	query := `
		UPDATE users
		SET role = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at
	`

	var updatedAt interface{}
	err := r.db.QueryRow(ctx, query, userID, newRole).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (r *UserRepository) GetAllUsers(ctx context.Context, limit, offset int, role string) ([]*models.User, error) {
	var query string
	var args []interface{}

	if role != "" {
		query = `
			SELECT id, username, email, password_hash, display_name, avatar_url,
			       bio, role, status, last_seen_at, created_at, updated_at
			FROM users
			WHERE deleted_at IS NULL AND role = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{role, limit, offset}
	} else {
		query = `
			SELECT id, username, email, password_hash, display_name, avatar_url,
			       bio, role, status, last_seen_at, created_at, updated_at
			FROM users
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.DisplayName,
			&user.AvatarURL,
			&user.Bio,
			&user.Role,
			&user.Status,
			&user.LastSeenAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) CountUsers(ctx context.Context, role string) (int64, error) {
	var query string
	var args []interface{}
	var count int64

	if role != "" {
		query = `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND role = $1`
		args = []interface{}{role}
	} else {
		query = `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
		args = []interface{}{}
	}

	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *UserRepository) SoftDeleteUser(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) GetUserStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(*) FILTER (WHERE role = 'admin') as admin_count,
			COUNT(*) FILTER (WHERE role = 'moderator') as moderator_count,
			COUNT(*) FILTER (WHERE role = 'user') as user_count,
			COUNT(*) FILTER (WHERE status = 'online') as online_count,
			COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') as new_users_week,
			COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '30 days') as new_users_month
		FROM users
		WHERE deleted_at IS NULL
	`

	var stats struct {
		TotalUsers     int64
		AdminCount     int64
		ModeratorCount int64
		UserCount      int64
		OnlineCount    int64
		NewUsersWeek   int64
		NewUsersMonth  int64
	}

	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalUsers,
		&stats.AdminCount,
		&stats.ModeratorCount,
		&stats.UserCount,
		&stats.OnlineCount,
		&stats.NewUsersWeek,
		&stats.NewUsersMonth,
	)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"total_users":     stats.TotalUsers,
		"admin_count":     stats.AdminCount,
		"moderator_count": stats.ModeratorCount,
		"user_count":      stats.UserCount,
		"online_count":    stats.OnlineCount,
		"new_users_week":  stats.NewUsersWeek,
		"new_users_month": stats.NewUsersMonth,
	}

	return result, nil
}
