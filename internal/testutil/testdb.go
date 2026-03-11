//go:build integration

package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	gophermart "github.com/iliaonishchenko/gophermart"
	"github.com/iliaonishchenko/gophermart/internal/auth"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	dbName = "gophermart_test"
	dbUser = "test"
	dbPass = "test"
)

type TestDB struct {
	DB        *sql.DB
	container *postgres.PostgresContainer
}

func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	if err := gophermart.RunMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	tdb := &TestDB{DB: db, container: container}

	t.Cleanup(func() {
		db.Close()
		container.Terminate(context.Background())
	})

	return tdb
}

func (tdb *TestDB) TruncateAll(t *testing.T) {
	t.Helper()
	_, err := tdb.DB.Exec("TRUNCATE TABLE withdrawals, balances, orders, users CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

func ContextWithUserUUID(uuid string) context.Context {
	return context.WithValue(context.Background(), auth.UserUUIDKey, uuid)
}

func (tdb *TestDB) CreateTestUser(t *testing.T, login, passwordHash string) *models.User {
	t.Helper()
	var id string
	var createdAt time.Time
	err := tdb.DB.QueryRow(
		"INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, created_at",
		login, passwordHash,
	).Scan(&id, &createdAt)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return &models.User{ID: &id, Login: login, PasswordHash: passwordHash, CreatedAt: createdAt}
}

func (tdb *TestDB) CreateTestOrder(t *testing.T, number, userID string, status models.OrderStatus, accrual *float32) *models.Order {
	t.Helper()
	var uploadedAt time.Time
	err := tdb.DB.QueryRow(
		"INSERT INTO orders (number, user_id, status, accrual) VALUES ($1, $2, $3, $4) RETURNING uploaded_at",
		number, userID, status, accrual,
	).Scan(&uploadedAt)
	if err != nil {
		t.Fatalf("failed to create test order: %v", err)
	}
	return &models.Order{Number: number, UserID: userID, Status: status, Accrual: accrual, UploadedAt: uploadedAt}
}

func (tdb *TestDB) SetBalance(t *testing.T, userID string, amount float32) {
	t.Helper()
	_, err := tdb.DB.Exec(
		fmt.Sprintf("INSERT INTO balances (user_id, balance) VALUES ('%s', %f) ON CONFLICT (user_id) DO UPDATE SET balance = %f", userID, amount, amount),
	)
	if err != nil {
		t.Fatalf("failed to set balance: %v", err)
	}
}
