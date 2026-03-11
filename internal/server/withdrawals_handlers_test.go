package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/internal/server/mocks"
	"github.com/iliaonishchenko/gophermart/pkg/api"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPostAPIUserBalanceWithdraw(t *testing.T) {
	t.Run("not enough funds returns 402", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockWithdrawals := mocks.NewMockWithdrawalService(ctrl)
		mockWithdrawals.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, models.ErrWithdrawalNotEnoughFunds)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mockWithdrawals, mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.PostAPIUserBalanceWithdraw(context.Background(), api.PostAPIUserBalanceWithdrawRequestObject{
			Body: &api.WithdrawalRequest{Order: "2377225624", Sum: 751},
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserBalanceWithdraw402Response{}, resp)
	})

	t.Run("non-existent order returns 422", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockWithdrawals := mocks.NewMockWithdrawalService(ctrl)
		mockWithdrawals.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, models.ErrWithdrawalNonExistentOrder)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mockWithdrawals, mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.PostAPIUserBalanceWithdraw(context.Background(), api.PostAPIUserBalanceWithdrawRequestObject{
			Body: &api.WithdrawalRequest{Order: "9999999999", Sum: 100},
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserBalanceWithdraw422Response{}, resp)
	})

	t.Run("service error returns 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockWithdrawals := mocks.NewMockWithdrawalService(ctrl)
		mockWithdrawals.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mockWithdrawals, mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.PostAPIUserBalanceWithdraw(context.Background(), api.PostAPIUserBalanceWithdrawRequestObject{
			Body: &api.WithdrawalRequest{Order: "2377225624", Sum: 500},
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserBalanceWithdraw500Response{}, resp)
	})

	t.Run("success returns 200", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockWithdrawals := mocks.NewMockWithdrawalService(ctrl)
		mockWithdrawals.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&models.Withdrawal{
			Order: "2377225624",
			Sum:   500,
		}, nil)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mockWithdrawals, mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.PostAPIUserBalanceWithdraw(context.Background(), api.PostAPIUserBalanceWithdrawRequestObject{
			Body: &api.WithdrawalRequest{Order: "2377225624", Sum: 500},
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserBalanceWithdraw200Response{}, resp)
	})
}

func TestGetAPIUserWithdrawals(t *testing.T) {
	t.Run("service error returns 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockWithdrawals := mocks.NewMockWithdrawalService(ctrl)
		mockWithdrawals.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db error"))

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mockWithdrawals, mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.GetAPIUserWithdrawals(context.Background(), api.GetAPIUserWithdrawalsRequestObject{})

		assert.NoError(t, err)
		assert.IsType(t, api.GetAPIUserWithdrawals500Response{}, resp)
	})

	t.Run("no withdrawals returns 204", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockWithdrawals := mocks.NewMockWithdrawalService(ctrl)
		mockWithdrawals.EXPECT().Get(gomock.Any()).Return([]*models.Withdrawal{}, nil)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mockWithdrawals, mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.GetAPIUserWithdrawals(context.Background(), api.GetAPIUserWithdrawalsRequestObject{})

		assert.NoError(t, err)
		assert.IsType(t, api.GetAPIUserWithdrawals204Response{}, resp)
	})

	t.Run("nil withdrawals returns 204", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockWithdrawals := mocks.NewMockWithdrawalService(ctrl)
		mockWithdrawals.EXPECT().Get(gomock.Any()).Return(nil, nil)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mockWithdrawals, mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.GetAPIUserWithdrawals(context.Background(), api.GetAPIUserWithdrawalsRequestObject{})

		assert.NoError(t, err)
		assert.IsType(t, api.GetAPIUserWithdrawals204Response{}, resp)
	})

	t.Run("success returns 200 with withdrawals", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		now := time.Now()
		mockWithdrawals := mocks.NewMockWithdrawalService(ctrl)
		mockWithdrawals.EXPECT().Get(gomock.Any()).Return([]*models.Withdrawal{
			{Order: "2377225624", Sum: 500, ProcessedAt: now},
		}, nil)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mockWithdrawals, mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.GetAPIUserWithdrawals(context.Background(), api.GetAPIUserWithdrawalsRequestObject{})

		assert.NoError(t, err)
		jsonResp, ok := resp.(api.GetAPIUserWithdrawals200JSONResponse)
		assert.True(t, ok)
		assert.Len(t, jsonResp, 1)
		assert.Equal(t, "2377225624", jsonResp[0].Order)
		assert.Equal(t, float32(500), jsonResp[0].Sum)
		assert.Equal(t, now.Format(time.RFC3339), jsonResp[0].ProcessedAt)
	})

	t.Run("multiple withdrawals returned in order", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		now := time.Now()
		earlier := now.Add(-1 * time.Hour)
		mockWithdrawals := mocks.NewMockWithdrawalService(ctrl)
		mockWithdrawals.EXPECT().Get(gomock.Any()).Return([]*models.Withdrawal{
			{Order: "1111111111", Sum: 300, ProcessedAt: now},
			{Order: "2222222222", Sum: 200, ProcessedAt: earlier},
		}, nil)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mockWithdrawals, mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.GetAPIUserWithdrawals(context.Background(), api.GetAPIUserWithdrawalsRequestObject{})

		assert.NoError(t, err)
		jsonResp, ok := resp.(api.GetAPIUserWithdrawals200JSONResponse)
		assert.True(t, ok)
		assert.Len(t, jsonResp, 2)
		assert.Equal(t, "1111111111", jsonResp[0].Order)
		assert.Equal(t, "2222222222", jsonResp[1].Order)
	})
}
