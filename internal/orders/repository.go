package orders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
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

func (r *Repository) Create(ctx context.Context, order *models.Order) (*models.Order, error) {
	query := `INSERT INTO orders (number, user_id, status, accrual) VALUES ($1, $2, $3, $4) RETURNING uploaded_at`

	if order == nil {
		return nil, errors.New("order is nil")
	}

	err := r.db.QueryRowContext(ctx, query, order.Number, order.UserID, order.Status, order.Accrual).Scan(&order.UploadedAt)
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
		var existingUserID string
		lookupQuery := `SELECT user_id FROM orders WHERE number = $1`
		if scanErr := r.db.QueryRowContext(ctx, lookupQuery, order.Number).Scan(&existingUserID); scanErr != nil {
			return nil, fmt.Errorf("failed to look up existing order: %w", scanErr)
		}
		if existingUserID == order.UserID {
			return nil, models.ErrOrderAlreadyExists
		}
		return nil, models.ErrOrderExistsDifferentUser
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

func (r *Repository) Get(ctx context.Context, userUUID *string) ([]*models.Order, error) {
	query := `SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`

	rows, err := r.db.QueryContext(ctx, query, *userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	defer rows.Close()
	var orders []*models.Order
	for rows.Next() {
		var order models.Order

		err := rows.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate orders: %w", err)
	}
	return orders, nil
}

func (r *Repository) Update(ctx context.Context, order *models.Order) (*models.Order, error) {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3 RETURNING number, user_id, status, accrual, uploaded_at`

	err := r.db.QueryRowContext(ctx, query, order.Status, order.Accrual, order.Number).Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}
	return order, nil
}
