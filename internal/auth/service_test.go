package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/gophermart/internal/auth/mocks"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hash)
}

func TestAuthenticate(t *testing.T) {
	tests := []struct {
		name      string
		login     string
		password  string
		mockSetup func(repo *mocks.MockUsersRepository, jwt *mocks.MockJwtGenerator)
		wantToken string
		wantErr   bool
	}{
		{
			name:     "successful authentication",
			login:    "testuser",
			password: "secret",
			mockSetup: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJwtGenerator) {
				user := &models.User{
					Login:        "testuser",
					PasswordHash: hashPassword(t, "secret"),
					CreatedAt:    time.Now(),
				}
				repo.EXPECT().GetUserByLogin(gomock.Any(), "testuser").Return(user, nil)
				jwt.EXPECT().GenerateToken("testuser").Return("jwt-token", nil)
			},
			wantToken: "jwt-token",
			wantErr:   false,
		},
		{
			name:     "user not found",
			login:    "unknown",
			password: "secret",
			mockSetup: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJwtGenerator) {
				repo.EXPECT().GetUserByLogin(gomock.Any(), "unknown").Return(nil, sql.ErrNoRows)
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name:     "wrong password",
			login:    "testuser",
			password: "wrongpassword",
			mockSetup: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJwtGenerator) {
				user := &models.User{
					Login:        "testuser",
					PasswordHash: hashPassword(t, "secret"),
					CreatedAt:    time.Now(),
				}
				repo.EXPECT().GetUserByLogin(gomock.Any(), "testuser").Return(user, nil)
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name:     "token generation fails",
			login:    "testuser",
			password: "secret",
			mockSetup: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJwtGenerator) {
				user := &models.User{
					Login:        "testuser",
					PasswordHash: hashPassword(t, "secret"),
					CreatedAt:    time.Now(),
				}
				repo.EXPECT().GetUserByLogin(gomock.Any(), "testuser").Return(user, nil)
				jwt.EXPECT().GenerateToken("testuser").Return("", errors.New("signing error"))
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockUsersRepository(ctrl)
			mockJwt := mocks.NewMockJwtGenerator(ctrl)
			tt.mockSetup(mockRepo, mockJwt)

			svc := NewAuthService(mockRepo, mockJwt)
			token, err := svc.Authenticate(context.Background(), tt.login, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}
