package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/iliaonishchenko/gophermart/internal/logger"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Ping() error {
	return r.db.Ping()
}

func (r *Repository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, created_at`

	if user == nil {
		return nil, errors.New("user is nil")
	}

	err := r.db.QueryRowContext(ctx, query, user.Login, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
		return nil, fmt.Errorf("%w: %s", models.ErrUserAlreadyExists, user.Login)
	}
	if err != nil {
		logger.Log.Error("failed to create user in database", logger.Err(err))
		return nil, err
	}

	return user, nil
}

func (r *Repository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`

	if login == "" {
		return nil, errors.New("login is nil")
	}

	var id string
	var uLogin string
	var passwordHash string
	var createdAt time.Time

	row := r.db.QueryRowContext(ctx, query, login)
	err := row.Scan(&id, &uLogin, &passwordHash, &createdAt)
	if err != nil {
		logger.Log.Error("failed to get user by login", logger.Err(err))
		return nil, err
	}
	user := &models.User{ID: &id, Login: uLogin, PasswordHash: passwordHash, CreatedAt: createdAt}
	return user, nil

}
