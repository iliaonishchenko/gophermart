//go:build integration

package balance

import (
	"context"
	"testing"

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

	t.Run("Get", func(t *testing.T) {
		t.Run("zero_balance_new_user", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "alice", "hash")
			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB, zap.NewNop())

			balance, err := repo.Get(ctx)

			require.NoError(t, err)
			require.NotNil(t, balance)
			assert.Equal(t, float32(0), balance.Current)
			assert.Equal(t, float32(0), balance.Withdrawn)
		})

		t.Run("balance_no_withdrawals", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "bob", "hash")
			tdb.SetBalance(t, *user.ID, 100)
			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB, zap.NewNop())

			balance, err := repo.Get(ctx)

			require.NoError(t, err)
			require.NotNil(t, balance)
			assert.InDelta(t, float32(100), balance.Current, 0.01)
			assert.Equal(t, float32(0), balance.Withdrawn)
		})

		t.Run("balance_with_withdrawals", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "charlie", "hash")
			tdb.CreateTestOrder(t, "100", *user.ID, models.OrderStatusProcessed, nil)
			tdb.CreateTestOrder(t, "200", *user.ID, models.OrderStatusProcessed, nil)
			tdb.SetBalance(t, *user.ID, 250)

			// Insert withdrawals directly
			_, err := tdb.DB.Exec("INSERT INTO withdrawals (order_number, user_id, sum) VALUES ($1, $2, $3)", "100", *user.ID, 30)
			require.NoError(t, err)
			_, err = tdb.DB.Exec("INSERT INTO withdrawals (order_number, user_id, sum) VALUES ($1, $2, $3)", "200", *user.ID, 20)
			require.NoError(t, err)

			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB, zap.NewNop())

			balance, err := repo.Get(ctx)

			require.NoError(t, err)
			require.NotNil(t, balance)
			assert.InDelta(t, float32(250), balance.Current, 0.01)
			assert.InDelta(t, float32(50), balance.Withdrawn, 0.01)
		})

		t.Run("no_user_in_context", func(t *testing.T) {
			repo := NewRepository(tdb.DB, zap.NewNop())

			_, err := repo.Get(context.Background())

			require.Error(t, err)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("inserts_new", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "alice", "hash")
			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB, zap.NewNop())

			amount := float32(100)
			err := repo.Update(ctx, &amount, *user.ID)

			require.NoError(t, err)

			balance, err := repo.Get(ctx)
			require.NoError(t, err)
			assert.InDelta(t, float32(100), balance.Current, 0.01)
		})

		t.Run("adds_to_existing", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "bob", "hash")
			tdb.SetBalance(t, *user.ID, 50)
			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB, zap.NewNop())

			amount := float32(75)
			err := repo.Update(ctx, &amount, *user.ID)

			require.NoError(t, err)

			balance, err := repo.Get(ctx)
			require.NoError(t, err)
			assert.InDelta(t, float32(125), balance.Current, 0.01)
		})

		t.Run("multiple_updates", func(t *testing.T) {
			tdb.TruncateAll(t)

			user := tdb.CreateTestUser(t, "charlie", "hash")
			ctx := testutil.ContextWithUserUUID(*user.ID)
			repo := NewRepository(tdb.DB, zap.NewNop())

			amounts := []float32{10, 20, 30}
			for _, a := range amounts {
				v := a
				err := repo.Update(ctx, &v, *user.ID)
				require.NoError(t, err)
			}

			balance, err := repo.Get(ctx)
			require.NoError(t, err)
			assert.InDelta(t, float32(60), balance.Current, 0.01)
		})
	})
}
