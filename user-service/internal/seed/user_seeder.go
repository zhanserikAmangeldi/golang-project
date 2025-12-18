package seed

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhanserikAmangeldi/user-service/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserSeeder struct {
	db *pgxpool.Pool
}

func NewUserSeeder(db *pgxpool.Pool) *UserSeeder {
	return &UserSeeder{db: db}
}

func (seeder *UserSeeder) SeedAdmin(ctx context.Context) error {
	password := "admin123"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &models.User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: string(hashed),
		Role:         "admin",
	}

	query := `
		INSERT INTO users (username, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err = seeder.db.QueryRow(ctx, query,
		admin.Username,
		admin.Email,
		admin.PasswordHash,
		admin.Role,
	).Scan(&admin.ID, &admin.CreatedAt, &admin.UpdatedAt)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_username_key\" (SQLSTATE 23505)" {
			return nil
		}
		return err
	}

	return nil
}
