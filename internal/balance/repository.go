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
	}
	if err != nil {
		logger.Log.Error("failed to scan balance", logger.Err(err))
		return nil, err
	}

	withdrawnQuery := "SELECT COALESCE(SUM(sum)) FROM withdrawals WHERE user_id = $1"
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
