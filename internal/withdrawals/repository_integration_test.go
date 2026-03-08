//go:build integration

package withdrawals

import (
	"context"
	"testing"
	"time"

	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryIntegration(t *testing.T) {
	tdb := testutil.NewTestDB(t)

	t.Run("Create", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "alice", "hash")
			tdb.CreateTestOrder(t, "12345678903", *user.ID, models.OrderStatusProcessed, nil)
			tdb.SetBalance(t, *user.ID, 500)

			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB)

			withdrawal := &models.Withdrawal{Order: "12345678903", Sum: 100}
			created, err := repo.Create(ctx, withdrawal)

			require.NoError(t, err)
			require.NotNil(t, created)
			assert.Equal(t, "12345678903", created.Order)
			assert.Equal(t, float32(100), created.Sum)
			assert.False(t, created.ProcessedAt.IsZero())
		})

		t.Run("insufficient_funds", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "bob", "hash")
			tdb.CreateTestOrder(t, "12345678903", *user.ID, models.OrderStatusProcessed, nil)
			tdb.SetBalance(t, *user.ID, 50)

			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB)

			withdrawal := &models.Withdrawal{Order: "12345678903", Sum: 100}
			_, err := repo.Create(ctx, withdrawal)

			require.Error(t, err)
			assert.ErrorIs(t, err, models.ErrWithdrawalNotEnoughFunds)
		})

		t.Run("zero_balance_no_row", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "charlie", "hash")
			tdb.CreateTestOrder(t, "12345678903", *user.ID, models.OrderStatusProcessed, nil)

			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB)

			withdrawal := &models.Withdrawal{Order: "12345678903", Sum: 10}
			_, err := repo.Create(ctx, withdrawal)

			require.Error(t, err)
			assert.ErrorIs(t, err, models.ErrWithdrawalNotEnoughFunds)
		})

		t.Run("nil_withdrawal", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "dave", "hash")
			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB)

			_, err := repo.Create(ctx, nil)

			require.Error(t, err)
		})

		t.Run("no_user_in_context", func(t *testing.T) {
			repo := NewRepository(tdb.DB)

			withdrawal := &models.Withdrawal{Order: "12345678903", Sum: 10}
			_, err := repo.Create(context.Background(), withdrawal)

			require.Error(t, err)
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("returns_withdrawals", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "alice", "hash")
			tdb.CreateTestOrder(t, "100", *user.ID, models.OrderStatusProcessed, nil)
			tdb.CreateTestOrder(t, "200", *user.ID, models.OrderStatusProcessed, nil)
			tdb.SetBalance(t, *user.ID, 1000)

			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB)

			w1 := &models.Withdrawal{Order: "100", Sum: 50}
			_, err := repo.Create(ctx, w1)
			require.NoError(t, err)

			time.Sleep(10 * time.Millisecond)

			w2 := &models.Withdrawal{Order: "200", Sum: 30}
			_, err = repo.Create(ctx, w2)
			require.NoError(t, err)

			withdrawals, err := repo.Get(ctx)

			require.NoError(t, err)
			require.Len(t, withdrawals, 2)
			assert.Equal(t, "200", withdrawals[0].Order)
			assert.Equal(t, "100", withdrawals[1].Order)
		})

		t.Run("empty", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "alice", "hash")
			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB)

			withdrawals, err := repo.Get(ctx)

			require.NoError(t, err)
			assert.Empty(t, withdrawals)
		})

		t.Run("no_user_in_context", func(t *testing.T) {
			repo := NewRepository(tdb.DB)

			_, err := repo.Get(context.Background())

			require.Error(t, err)
		})
	})
}
