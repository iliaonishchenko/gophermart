package balance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/iliaonishchenko/gophermart/internal/auth"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"go.uber.org/zap"
)

type Repository struct {
	db  *sql.DB
	log *zap.Logger
}

func NewRepository(db *sql.DB, log *zap.Logger) *Repository {
	return &Repository{db: db, log: log}
}

func (r *Repository) Ping() error {
	return r.db.Ping()
}

func (r *Repository) Get(ctx context.Context) (*models.Balance, error) {
	userUUID, err := auth.GetUserUUID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user uuid: %w", err)
	}
	var balance models.Balance
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	balanceQuery := "SELECT balance FROM balances WHERE user_id = $1"
	err = tx.QueryRowContext(ctx, balanceQuery, userUUID).Scan(&balance.Current)
	if errors.Is(err, sql.ErrNoRows) {
		balance.Current = 0
	} else if err != nil {
		return nil, err
	}

	withdrawnQuery := "SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1"
	err = tx.QueryRowContext(ctx, withdrawnQuery, userUUID).Scan(&balance.Withdrawn)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

func (r *Repository) Update(ctx context.Context, balance *float32, userUUID string) error {
	query := `INSERT INTO balances (user_id, balance) VALUES ($2, $1)
		ON CONFLICT (user_id) DO UPDATE SET balance = balances.balance + $1`

	_, err := r.db.ExecContext(ctx, query, balance, userUUID)
	if err != nil {
		return err
	}
	return nil
}
