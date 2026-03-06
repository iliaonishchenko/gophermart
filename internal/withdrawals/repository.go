package withdrawals

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/iliaonishchenko/gophermart/internal/auth"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
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

func (r *Repository) Create(ctx context.Context, withdrawalToCreate *models.Withdrawal) (*models.Withdrawal, error) {
	query := "INSERT INTO withdrawals (order_number, user_id, sum) VALUES ($1, $2, $3) RETURNING processed_at"
	userUUID, ok := auth.GetUserUUID(ctx)
	if !ok {
		return nil, errors.New("could not get user uuid")
	}

	if withdrawalToCreate == nil {
		return nil, errors.New("withdrawal is nil")
	}

	err := r.db.QueryRowContext(ctx, query, withdrawalToCreate.Order, userUUID, withdrawalToCreate.Sum).Scan(&withdrawalToCreate.ProcessedAt)

	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
		return nil, models.ErrWithdrawalNonExistentOrder
	}

	// todo check if user has enough funds to perform the operation

	if err != nil {
		return nil, fmt.Errorf("failed to insert withdrawal: %w", err)
	}

	return withdrawalToCreate, nil
}

func (r *Repository) Get(ctx context.Context) ([]*models.Withdrawal, error) {
	query := "SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC"
	userUUID, ok := auth.GetUserUUID(ctx)
	if !ok {
		return nil, errors.New("could not get user uuid")
	}
	rows, err := r.db.QueryContext(ctx, query, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}
	defer rows.Close()
	var withdrawals []*models.Withdrawal
	for rows.Next() {
		var withdrawal models.Withdrawal
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, &withdrawal)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate withdrawals: %w", err)
	}

	return withdrawals, nil
}
