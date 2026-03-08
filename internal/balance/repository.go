package balance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/iliaonishchenko/gophermart/internal/auth"
	"github.com/iliaonishchenko/gophermart/internal/logger"
	"github.com/iliaonishchenko/gophermart/internal/models"
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

func (r *Repository) Get(ctx context.Context) (*models.Balance, error) {
	userUUID, ok := auth.GetUserUUID(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get user uuid")
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
		logger.Log.Error("failed to scan balance", logger.Err(err))
		return nil, err
	}

	withdrawnQuery := "SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1"
	err = tx.QueryRowContext(ctx, withdrawnQuery, userUUID).Scan(&balance.Withdrawn)
	if err != nil {
		logger.Log.Error("failed to get all users withdrawals for balance", logger.Err(err))
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		logger.Log.Error("balance transaction failed", logger.Err(err))
		return nil, err
	}

	return &balance, nil
}

func (r *Repository) Update(ctx context.Context, balance *float32, userUUID string) error {
	query := `INSERT INTO balances (user_id, balance) VALUES ($2, $1)
		ON CONFLICT (user_id) DO UPDATE SET balance = balances.balance + $1`

	_, err := r.db.ExecContext(ctx, query, balance, userUUID)
	if err != nil {
		logger.Log.Error("failed to update balance", logger.Err(err))
		return err
	}
	return nil
}
