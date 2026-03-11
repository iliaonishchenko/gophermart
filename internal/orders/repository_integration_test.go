//go:build integration

package orders

import (
	"context"
	"testing"
	"time"

	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func float32Ptr(f float32) *float32 {
	return &f
}

func TestRepositoryIntegration(t *testing.T) {
	tdb := testutil.NewTestDB(t)

	t.Run("Create", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "alice", "hash")
			repo := NewRepository(tdb.DB, zap.NewNop())

			order := &models.Order{
				Number: "12345678903",
				UserID: *user.ID,
				Status: models.OrderStatusNew,
			}
			created, err := repo.Create(context.Background(), order)

			require.NoError(t, err)
			require.NotNil(t, created)
			assert.Equal(t, "12345678903", created.Number)
			assert.False(t, created.UploadedAt.IsZero())
		})

		t.Run("duplicate_same_user", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "bob", "hash")
			repo := NewRepository(tdb.DB, zap.NewNop())

			order1 := &models.Order{Number: "12345678903", UserID: *user.ID, Status: models.OrderStatusNew}
			_, err := repo.Create(context.Background(), order1)
			require.NoError(t, err)

			order2 := &models.Order{Number: "12345678903", UserID: *user.ID, Status: models.OrderStatusNew}
			_, err = repo.Create(context.Background(), order2)

			require.Error(t, err)
			assert.ErrorIs(t, err, models.ErrOrderAlreadyExists)
		})

		t.Run("duplicate_different_user", func(t *testing.T) {
			tdb.TruncateAll(t)

			user1 := tdb.CreateTestUser(t, "charlie", "hash")
			user2 := tdb.CreateTestUser(t, "dave", "hash")
			repo := NewRepository(tdb.DB, zap.NewNop())

			order1 := &models.Order{Number: "12345678903", UserID: *user1.ID, Status: models.OrderStatusNew}
			_, err := repo.Create(context.Background(), order1)
			require.NoError(t, err)

			order2 := &models.Order{Number: "12345678903", UserID: *user2.ID, Status: models.OrderStatusNew}
			_, err = repo.Create(context.Background(), order2)

			require.Error(t, err)
			assert.ErrorIs(t, err, models.ErrOrderExistsDifferentUser)
		})

		t.Run("nil_order", func(t *testing.T) {
			repo := NewRepository(tdb.DB, zap.NewNop())

			_, err := repo.Create(context.Background(), nil)

			require.Error(t, err)
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("returns_user_orders", func(t *testing.T) {
			tdb.TruncateAll(t)

			user1 := tdb.CreateTestUser(t, "alice", "hash")
			user2 := tdb.CreateTestUser(t, "bob", "hash")
			repo := NewRepository(tdb.DB, zap.NewNop())

			o1 := &models.Order{Number: "100", UserID: *user1.ID, Status: models.OrderStatusNew}
			_, err := repo.Create(context.Background(), o1)
			require.NoError(t, err)

			time.Sleep(10 * time.Millisecond)

			o2 := &models.Order{Number: "200", UserID: *user1.ID, Status: models.OrderStatusProcessed, Accrual: float32Ptr(50)}
			_, err = repo.Create(context.Background(), o2)
			require.NoError(t, err)

			o3 := &models.Order{Number: "300", UserID: *user2.ID, Status: models.OrderStatusNew}
			_, err = repo.Create(context.Background(), o3)
			require.NoError(t, err)

			orders, err := repo.Get(context.Background(), user1.ID)

			require.NoError(t, err)
			require.Len(t, orders, 2)
			assert.Equal(t, "200", orders[0].Number)
			assert.Equal(t, "100", orders[1].Number)
		})

		t.Run("empty_for_user_without_orders", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "alice", "hash")
			repo := NewRepository(tdb.DB, zap.NewNop())

			orders, err := repo.Get(context.Background(), user.ID)

			require.NoError(t, err)
			assert.Empty(t, orders)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "alice", "hash")
			repo := NewRepository(tdb.DB, zap.NewNop())

			order := &models.Order{Number: "12345678903", UserID: *user.ID, Status: models.OrderStatusNew}
			_, err := repo.Create(context.Background(), order)
			require.NoError(t, err)

			accrual := float32Ptr(100.50)
			updateOrder := &models.Order{
				Number:  "12345678903",
				Status:  models.OrderStatusProcessed,
				Accrual: accrual,
			}
			updated, err := repo.Update(context.Background(), updateOrder)

			require.NoError(t, err)
			require.NotNil(t, updated)
			assert.Equal(t, models.OrderStatusProcessed, updated.Status)
			require.NotNil(t, updated.Accrual)
			assert.InDelta(t, float32(100.50), *updated.Accrual, 0.01)
			assert.Equal(t, *user.ID, updated.UserID)
		})
	})
}
