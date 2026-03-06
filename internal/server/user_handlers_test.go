package server

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/internal/server/mocks"
	"github.com/iliaonishchenko/gophermart/pkg/api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPostAPIUserRegister(t *testing.T) {
	t.Run("empty login returns 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserRegister(context.Background(), api.PostAPIUserRegisterRequestObject{
			Body: &api.UserCredentials{Login: "", Password: "pass"},
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserRegister400Response{}, resp)
	})

	t.Run("empty password returns 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserRegister(context.Background(), api.PostAPIUserRegisterRequestObject{
			Body: &api.UserCredentials{Login: "user", Password: ""},
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserRegister400Response{}, resp)
	})

	t.Run("hash password error returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuth := mocks.NewMockAuth(ctrl)
		mockAuth.EXPECT().GeneratePasswordHash("pass").Return("", errors.New("hash error"))

		srv := NewServer(mockAuth, mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserRegister(context.Background(), api.PostAPIUserRegisterRequestObject{
			Body: &api.UserCredentials{Login: "user", Password: "pass"},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("user already exists returns 409", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuth := mocks.NewMockAuth(ctrl)
		mockAuth.EXPECT().GeneratePasswordHash("pass").Return("hashed", nil)

		mockUsers := mocks.NewMockUserService(ctrl)
		mockUsers.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, models.ErrUserAlreadyExists)

		srv := NewServer(mockAuth, mockUsers, mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserRegister(context.Background(), api.PostAPIUserRegisterRequestObject{
			Body: &api.UserCredentials{Login: "user", Password: "pass"},
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserRegister409Response{}, resp)
	})

	t.Run("create user error returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuth := mocks.NewMockAuth(ctrl)
		mockAuth.EXPECT().GeneratePasswordHash("pass").Return("hashed", nil)

		mockUsers := mocks.NewMockUserService(ctrl)
		mockUsers.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

		srv := NewServer(mockAuth, mockUsers, mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserRegister(context.Background(), api.PostAPIUserRegisterRequestObject{
			Body: &api.UserCredentials{Login: "user", Password: "pass"},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("generate token error returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userID := "some-uuid"
		mockAuth := mocks.NewMockAuth(ctrl)
		mockAuth.EXPECT().GeneratePasswordHash("pass").Return("hashed", nil)
		mockAuth.EXPECT().GenerateToken(&userID).Return("", errors.New("token error"))

		mockUsers := mocks.NewMockUserService(ctrl)
		mockUsers.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&models.User{ID: &userID}, nil)

		srv := NewServer(mockAuth, mockUsers, mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserRegister(context.Background(), api.PostAPIUserRegisterRequestObject{
			Body: &api.UserCredentials{Login: "user", Password: "pass"},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("success returns 200 with token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userID := "some-uuid"
		mockAuth := mocks.NewMockAuth(ctrl)
		mockAuth.EXPECT().GeneratePasswordHash("pass").Return("hashed", nil)
		mockAuth.EXPECT().GenerateToken(&userID).Return("jwt-token", nil)

		mockUsers := mocks.NewMockUserService(ctrl)
		mockUsers.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&models.User{ID: &userID}, nil)

		srv := NewServer(mockAuth, mockUsers, mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserRegister(context.Background(), api.PostAPIUserRegisterRequestObject{
			Body: &api.UserCredentials{Login: "user", Password: "pass"},
		})

		assert.NoError(t, err)
		r, ok := resp.(api.PostAPIUserRegister200Response)
		assert.True(t, ok)
		assert.Equal(t, "Bearer jwt-token", r.Headers.Authorization)
	})
}

func TestPostAPIUserLogin(t *testing.T) {
	t.Run("empty login returns 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserLogin(context.Background(), api.PostAPIUserLoginRequestObject{
			Body: &api.UserCredentials{Login: "", Password: "pass"},
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserLogin400Response{}, resp)
	})

	t.Run("empty password returns 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserLogin(context.Background(), api.PostAPIUserLoginRequestObject{
			Body: &api.UserCredentials{Login: "user", Password: ""},
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserLogin400Response{}, resp)
	})

	t.Run("invalid credentials returns 401", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuth := mocks.NewMockAuth(ctrl)
		mockAuth.EXPECT().Authenticate(gomock.Any(), "user", "wrong").Return("", models.ErrInvalidCredentials)

		srv := NewServer(mockAuth, mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserLogin(context.Background(), api.PostAPIUserLoginRequestObject{
			Body: &api.UserCredentials{Login: "user", Password: "wrong"},
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserLogin401Response{}, resp)
	})

	t.Run("authenticate error returns 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuth := mocks.NewMockAuth(ctrl)
		mockAuth.EXPECT().Authenticate(gomock.Any(), "user", "pass").Return("", errors.New("db error"))

		srv := NewServer(mockAuth, mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserLogin(context.Background(), api.PostAPIUserLoginRequestObject{
			Body: &api.UserCredentials{Login: "user", Password: "pass"},
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserLogin500Response{}, resp)
	})

	t.Run("success returns 200 with token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuth := mocks.NewMockAuth(ctrl)
		mockAuth.EXPECT().Authenticate(gomock.Any(), "user", "pass").Return("jwt-token", nil)

		srv := NewServer(mockAuth, mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl))

		resp, err := srv.PostAPIUserLogin(context.Background(), api.PostAPIUserLoginRequestObject{
			Body: &api.UserCredentials{Login: "user", Password: "pass"},
		})

		assert.NoError(t, err)
		r, ok := resp.(api.PostAPIUserLogin200Response)
		assert.True(t, ok)
		assert.Equal(t, "Bearer jwt-token", r.Headers.Authorization)
	})
}
