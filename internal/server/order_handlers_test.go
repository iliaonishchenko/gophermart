package server

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/gophermart/internal/auth"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/internal/server/mocks"
	"github.com/iliaonishchenko/gophermart/pkg/api"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"
)

func ctxWithUUID(uuid string) context.Context {
	return context.WithValue(context.Background(), auth.UserUUIDKey, uuid)
}

func strPtr(s string) *string {
	return &s
}

func TestPostAPIUserOrders(t *testing.T) {
	t.Run("missing user UUID returns 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mocks.NewMockOrderService(ctrl), mocks.NewMockWithdrawalService(ctrl), mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		body := "12345678903"
		resp, err := srv.PostAPIUserOrders(context.Background(), api.PostAPIUserOrdersRequestObject{
			Body: &body,
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserOrders400Response{}, resp)
	})

	t.Run("invalid order format returns 422", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockOrders := mocks.NewMockOrderService(ctrl)
		mockOrders.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, models.ErrInvalidOrderFormat)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mockOrders, mocks.NewMockWithdrawalService(ctrl), mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		body := "12345678901"
		resp, err := srv.PostAPIUserOrders(ctxWithUUID("user-uuid"), api.PostAPIUserOrdersRequestObject{
			Body: &body,
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserOrders422Response{}, resp)
	})

	t.Run("order already exists by same user returns 200", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockOrders := mocks.NewMockOrderService(ctrl)
		mockOrders.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, models.ErrOrderAlreadyExists)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mockOrders, mocks.NewMockWithdrawalService(ctrl), mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		body := "12345678903"
		resp, err := srv.PostAPIUserOrders(ctxWithUUID("user-uuid"), api.PostAPIUserOrdersRequestObject{
			Body: &body,
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserOrders200Response{}, resp)
	})

	t.Run("order exists by different user returns 409", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockOrders := mocks.NewMockOrderService(ctrl)
		mockOrders.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, models.ErrOrderExistsDifferentUser)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mockOrders, mocks.NewMockWithdrawalService(ctrl), mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		body := "12345678903"
		resp, err := srv.PostAPIUserOrders(ctxWithUUID("user-uuid"), api.PostAPIUserOrdersRequestObject{
			Body: &body,
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserOrders409Response{}, resp)
	})

	t.Run("service error returns 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockOrders := mocks.NewMockOrderService(ctrl)
		mockOrders.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mockOrders, mocks.NewMockWithdrawalService(ctrl), mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		body := "12345678903"
		resp, err := srv.PostAPIUserOrders(ctxWithUUID("user-uuid"), api.PostAPIUserOrdersRequestObject{
			Body: &body,
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserOrders500Response{}, resp)
	})

	t.Run("success returns 202", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockOrders := mocks.NewMockOrderService(ctrl)
		mockOrders.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&models.Order{Number: "12345678903"}, nil)

		mockAccrual := mocks.NewMockAccrualService(ctrl)
		mockAccrual.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any())

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mockOrders, mocks.NewMockWithdrawalService(ctrl), mocks.NewMockBalanceService(ctrl), mockAccrual, zap.NewNop())

		body := "12345678903"
		resp, err := srv.PostAPIUserOrders(ctxWithUUID("user-uuid"), api.PostAPIUserOrdersRequestObject{
			Body: &body,
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserOrders202Response{}, resp)
	})

	t.Run("trims whitespace from order number", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockOrders := mocks.NewMockOrderService(ctrl)
		mockOrders.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, order *models.Order) (*models.Order, error) {
				assert.Equal(t, "12345678903", order.Number)
				return order, nil
			})

		mockAccrual := mocks.NewMockAccrualService(ctrl)
		mockAccrual.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any())

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mockOrders, mocks.NewMockWithdrawalService(ctrl), mocks.NewMockBalanceService(ctrl), mockAccrual, zap.NewNop())

		body := "  12345678903\n"
		resp, err := srv.PostAPIUserOrders(ctxWithUUID("user-uuid"), api.PostAPIUserOrdersRequestObject{
			Body: &body,
		})

		assert.NoError(t, err)
		assert.IsType(t, api.PostAPIUserOrders202Response{}, resp)
	})
}

func TestGetAPIUserOrders(t *testing.T) {
	t.Run("service error returns 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockOrders := mocks.NewMockOrderService(ctrl)
		mockOrders.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mockOrders, mocks.NewMockWithdrawalService(ctrl), mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.GetAPIUserOrders(ctxWithUUID("user-uuid"), api.GetAPIUserOrdersRequestObject{})

		assert.NoError(t, err)
		assert.IsType(t, api.GetAPIUserOrders500Response{}, resp)
	})

	t.Run("no orders returns 204", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockOrders := mocks.NewMockOrderService(ctrl)
		mockOrders.EXPECT().Get(gomock.Any(), gomock.Any()).Return([]*models.Order{}, nil)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mockOrders, mocks.NewMockWithdrawalService(ctrl), mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.GetAPIUserOrders(ctxWithUUID("user-uuid"), api.GetAPIUserOrdersRequestObject{})

		assert.NoError(t, err)
		assert.IsType(t, api.GetAPIUserOrders204Response{}, resp)
	})

	t.Run("success returns 200 with orders", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		now := time.Now()
		var accrual float32 = 500
		mockOrders := mocks.NewMockOrderService(ctrl)
		mockOrders.EXPECT().Get(gomock.Any(), gomock.Any()).Return([]*models.Order{
			{Number: "12345678903", UserID: "user-uuid", Status: models.OrderStatusProcessed, Accrual: &accrual, UploadedAt: now},
			{Number: "9278923470", UserID: "user-uuid", Status: models.OrderStatusNew, Accrual: nil, UploadedAt: now},
		}, nil)

		srv := NewServer(mocks.NewMockAuth(ctrl), mocks.NewMockUserService(ctrl), mockOrders, mocks.NewMockWithdrawalService(ctrl), mocks.NewMockBalanceService(ctrl), mocks.NewMockAccrualService(ctrl), zap.NewNop())

		resp, err := srv.GetAPIUserOrders(ctxWithUUID("user-uuid"), api.GetAPIUserOrdersRequestObject{})

		assert.NoError(t, err)
		orders, ok := resp.(api.GetAPIUserOrders200JSONResponse)
		assert.True(t, ok)
		assert.Len(t, orders, 2)
		assert.Equal(t, "12345678903", orders[0].Number)
		assert.Equal(t, api.OrderStatus("PROCESSED"), orders[0].Status)
		assert.Equal(t, &accrual, orders[0].Accrual)
		assert.Equal(t, "9278923470", orders[1].Number)
		assert.Nil(t, orders[1].Accrual)
	})
}
