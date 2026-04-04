package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

type UserRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewUserRepository(db *pgxpool.Pool, log *slog.Logger) *UserRepository {
	return &UserRepository{db: db, log: log}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, email, password_hash, role, created_at)
	          VALUES ($1, $2, $3, $4, $5)`

	id := uuid.New()
	now := time.Now().UTC()

	_, err := r.db.Exec(ctx, query, id, user.Email, user.PasswordHash, user.Role, now)
	if err != nil {
		r.log.Error("failed to create user", "error", err)
		return err
	}

	user.ID = id
	user.CreatedAt = now

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`

	var u model.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		r.log.Error("failed to get user by email", "error", err)
		return nil, err
	}

	return &u, nil
}
