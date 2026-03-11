package accrual

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/gophermart/internal/accrual/mocks"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func float32PtrPoller(f float32) *float32 {
	return &f
}

func TestHandleOrder(t *testing.T) {
	tests := []struct {
		name    string
		order   *order
		setup   func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService)
		wantErr bool
	}{
		{
			name:  "client error returns error",
			order: &order{number: "123", status: "REGISTERED", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				client.EXPECT().Evaluate("123").Return(nil, errors.New("connection refused"))
			},
			wantErr: true,
		},
		{
			name:  "status unchanged REGISTERED→REGISTERED returns errOrderNotReady",
			order: &order{number: "123", status: "REGISTERED", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				client.EXPECT().Evaluate("123").Return(&models.AccrualOrder{
					Order:  "123",
					Status: models.StatusRegistered,
				}, nil)
			},
			wantErr: true,
		},
		{
			name:  "PROCESSING calls oService.Update returns errOrderNotReady",
			order: &order{number: "123", status: "REGISTERED", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				client.EXPECT().Evaluate("123").Return(&models.AccrualOrder{
					Order:  "123",
					Status: models.StatusProcessing,
				}, nil)
				oSvc.EXPECT().Update(gomock.Any(), &models.Order{
					Number: "123",
					Status: models.OrderStatusProcessing,
				}).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:  "PROCESSING with oService.Update error returns error",
			order: &order{number: "123", status: "REGISTERED", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				client.EXPECT().Evaluate("123").Return(&models.AccrualOrder{
					Order:  "123",
					Status: models.StatusProcessing,
				}, nil)
				oSvc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:  "PROCESSED calls oService.Update and bService.Update returns nil",
			order: &order{number: "123", status: "PROCESSING", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				accrual := float32PtrPoller(100.5)
				client.EXPECT().Evaluate("123").Return(&models.AccrualOrder{
					Order:   "123",
					Status:  models.StatusProcessed,
					Accrual: accrual,
				}, nil)
				oSvc.EXPECT().Update(gomock.Any(), &models.Order{
					Number:  "123",
					Status:  models.OrderStatusProcessed,
					Accrual: accrual,
				}).Return(nil, nil)
				bSvc.EXPECT().Update(gomock.Any(), accrual, "user-1").Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "PROCESSED with nil accrual skips bService.Update returns nil",
			order: &order{number: "123", status: "PROCESSING", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				client.EXPECT().Evaluate("123").Return(&models.AccrualOrder{
					Order:  "123",
					Status: models.StatusProcessed,
				}, nil)
				oSvc.EXPECT().Update(gomock.Any(), &models.Order{
					Number: "123",
					Status: models.OrderStatusProcessed,
				}).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name:  "INVALID calls oService.Update skips bService.Update returns nil",
			order: &order{number: "123", status: "PROCESSING", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				client.EXPECT().Evaluate("123").Return(&models.AccrualOrder{
					Order:  "123",
					Status: models.StatusInvalid,
				}, nil)
				oSvc.EXPECT().Update(gomock.Any(), &models.Order{
					Number: "123",
					Status: models.OrderStatusInvalid,
				}).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name:  "oService.Update error on final status returns error",
			order: &order{number: "123", status: "PROCESSING", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				client.EXPECT().Evaluate("123").Return(&models.AccrualOrder{
					Order:   "123",
					Status:  models.StatusProcessed,
					Accrual: float32PtrPoller(50),
				}, nil)
				oSvc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:  "bService.Update error returns error",
			order: &order{number: "123", status: "PROCESSING", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				accrual := float32PtrPoller(50)
				client.EXPECT().Evaluate("123").Return(&models.AccrualOrder{
					Order:   "123",
					Status:  models.StatusProcessed,
					Accrual: accrual,
				}, nil)
				oSvc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, nil)
				bSvc.EXPECT().Update(gomock.Any(), accrual, "user-1").Return(errors.New("balance error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockAccrualClient(ctrl)
			mockOSvc := mocks.NewMockOrderService(ctrl)
			mockBSvc := mocks.NewMockBalanceService(ctrl)

			tt.setup(mockClient, mockOSvc, mockBSvc)

			poller := NewAccrualPoller(mockClient, mockOSvc, mockBSvc, zap.NewNop())
			err := poller.handleOrder(context.Background(), tt.order)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	poller := NewAccrualPoller(
		mocks.NewMockAccrualClient(ctrl),
		mocks.NewMockOrderService(ctrl),
		mocks.NewMockBalanceService(ctrl),
		zap.NewNop(),
	)

	poller.Add("12345", "NEW", "user-uuid")

	assert.Len(t, poller.orders, 1)
	assert.Equal(t, "12345", poller.orders[0].number)
	assert.Equal(t, "NEW", poller.orders[0].status)
	assert.Equal(t, "user-uuid", poller.orders[0].userUUID)
}

func TestRun_GlobalPauseOnRateLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAccrualClient(ctrl)
	mockOSvc := mocks.NewMockOrderService(ctrl)
	mockBSvc := mocks.NewMockBalanceService(ctrl)

	var callCount int32

	mockClient.EXPECT().Evaluate(gomock.Any()).DoAndReturn(func(number string) (*models.AccrualOrder, error) {
		n := atomic.AddInt32(&callCount, 1)
		if n <= 3 {
			return nil, &RateLimitError{RetryAfter: 200 * time.Millisecond}
		}
		return &models.AccrualOrder{
			Order:  number,
			Status: models.StatusProcessed,
		}, nil
	}).AnyTimes()

	mockOSvc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	poller := NewAccrualPoller(mockClient, mockOSvc, mockBSvc, zap.NewNop())
	poller.ticker = time.NewTicker(50 * time.Millisecond)

	poller.Add("order-1", "REGISTERED", "user-1")
	poller.Add("order-2", "REGISTERED", "user-2")
	poller.Add("order-3", "REGISTERED", "user-3")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		poller.Run(ctx)
		close(done)
	}()

	<-done

	poller.mu.RLock()
	remaining := len(poller.orders)
	poller.mu.RUnlock()
	assert.Equal(t, 0, remaining, "all orders should be processed after pause expires")

	total := atomic.LoadInt32(&callCount)
	assert.Greater(t, total, int32(3), "should have made calls after pause expired")
}
