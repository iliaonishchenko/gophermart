//go:build integration

package users

import (
	"context"
	"testing"

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

			repo := NewRepository(tdb.DB)
			user := &models.User{Login: "alice", PasswordHash: "hash123"}

			created, err := repo.Create(context.Background(), user)

			require.NoError(t, err)
			require.NotNil(t, created)
			assert.NotNil(t, created.ID)
			assert.NotEmpty(t, *created.ID)
			assert.Equal(t, "alice", created.Login)
			assert.False(t, created.CreatedAt.IsZero())
		})

		t.Run("duplicate_login", func(t *testing.T) {
			tdb.TruncateAll(t)

			repo := NewRepository(tdb.DB)
			user1 := &models.User{Login: "bob", PasswordHash: "hash1"}
			_, err := repo.Create(context.Background(), user1)
			require.NoError(t, err)

			user2 := &models.User{Login: "bob", PasswordHash: "hash2"}
			_, err = repo.Create(context.Background(), user2)

			require.Error(t, err)
			assert.ErrorIs(t, err, models.ErrUserAlreadyExists)
		})

		t.Run("nil_user", func(t *testing.T) {
			repo := NewRepository(tdb.DB)

			_, err := repo.Create(context.Background(), nil)

			require.Error(t, err)
		})
	})

	t.Run("GetUserByLogin", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			tdb.TruncateAll(t)

			repo := NewRepository(tdb.DB)
			user := &models.User{Login: "charlie", PasswordHash: "hash456"}
			created, err := repo.Create(context.Background(), user)
			require.NoError(t, err)

			found, err := repo.GetUserByLogin(context.Background(), "charlie")

			require.NoError(t, err)
			require.NotNil(t, found)
			assert.Equal(t, *created.ID, *found.ID)
			assert.Equal(t, "charlie", found.Login)
			assert.Equal(t, "hash456", found.PasswordHash)
			assert.False(t, found.CreatedAt.IsZero())
		})

		t.Run("not_found", func(t *testing.T) {
			tdb.TruncateAll(t)

			repo := NewRepository(tdb.DB)

			_, err := repo.GetUserByLogin(context.Background(), "nonexistent")

			require.Error(t, err)
		})

		t.Run("empty_login", func(t *testing.T) {
			repo := NewRepository(tdb.DB)

			_, err := repo.GetUserByLogin(context.Background(), "")

			require.Error(t, err)
		})
	})
}
