package server

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/internal/server/mocks"
	"github.com/iliaonishchenko/gophermart/pkg/api"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetAPIUserBalance(t *testing.T) {
	t.Run("service error returns 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBalance := mocks.NewMockBalanceService(ctrl)
		mockBalance.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db error"))

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl), mockBalance, mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.GetAPIUserBalance(context.Background(), api.GetAPIUserBalanceRequestObject{})

		assert.NoError(t, err)
		assert.IsType(t, api.GetAPIUserBalance500Response{}, resp)
	})

	t.Run("success returns 200 with balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBalance := mocks.NewMockBalanceService(ctrl)
		mockBalance.EXPECT().Get(gomock.Any()).Return(&models.Balance{Current: 500.5, Withdrawn: 42}, nil)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl), mockBalance, mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.GetAPIUserBalance(context.Background(), api.GetAPIUserBalanceRequestObject{})

		assert.NoError(t, err)
		jsonResp, ok := resp.(api.GetAPIUserBalance200JSONResponse)
		assert.True(t, ok)
		assert.Equal(t, float32(500.5), jsonResp.Current)
		assert.Equal(t, float32(42), jsonResp.Withdrawn)
	})

	t.Run("zero balance returns 200", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBalance := mocks.NewMockBalanceService(ctrl)
		mockBalance.EXPECT().Get(gomock.Any()).Return(&models.Balance{Current: 0, Withdrawn: 0}, nil)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl), mockBalance, mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.GetAPIUserBalance(context.Background(), api.GetAPIUserBalanceRequestObject{})

		assert.NoError(t, err)
		jsonResp, ok := resp.(api.GetAPIUserBalance200JSONResponse)
		assert.True(t, ok)
		assert.Equal(t, float32(0), jsonResp.Current)
		assert.Equal(t, float32(0), jsonResp.Withdrawn)
	})
}
