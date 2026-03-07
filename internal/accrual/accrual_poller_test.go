package accrual

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/gophermart/internal/accrual/mocks"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
)

func float32PtrPoller(f float32) *float32 {
	return &f
}

func TestHandleOrder(t *testing.T) {
	tests := []struct {
		name     string
		order    *order
		setup    func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService)
		wantDone bool
	}{
		{
			name:  "client error returns false",
			order: &order{number: "123", status: "REGISTERED", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				client.EXPECT().Evaluate("123").Return(nil, errors.New("connection refused"))
			},
			wantDone: false,
		},
		{
			name:  "status unchanged REGISTERED→REGISTERED returns false",
			order: &order{number: "123", status: "REGISTERED", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				client.EXPECT().Evaluate("123").Return(&models.AccrualOrder{
					Order:  "123",
					Status: models.StatusRegistered,
				}, nil)
			},
			wantDone: false,
		},
		{
			name:  "PROCESSING calls oService.Update returns false",
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
			wantDone: false,
		},
		{
			name:  "PROCESSING with oService.Update error still returns false",
			order: &order{number: "123", status: "REGISTERED", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				client.EXPECT().Evaluate("123").Return(&models.AccrualOrder{
					Order:  "123",
					Status: models.StatusProcessing,
				}, nil)
				oSvc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantDone: false,
		},
		{
			name:  "PROCESSED calls oService.Update and bService.Update returns true",
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
			wantDone: true,
		},
		{
			name:  "PROCESSED with nil accrual skips bService.Update returns true",
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
			wantDone: true,
		},
		{
			name:  "INVALID calls oService.Update skips bService.Update returns true",
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
			wantDone: true,
		},
		{
			name:  "oService.Update error on final status returns false",
			order: &order{number: "123", status: "PROCESSING", userUUID: "user-1"},
			setup: func(client *mocks.MockAccrualClient, oSvc *mocks.MockOrderService, bSvc *mocks.MockBalanceService) {
				client.EXPECT().Evaluate("123").Return(&models.AccrualOrder{
					Order:   "123",
					Status:  models.StatusProcessed,
					Accrual: float32PtrPoller(50),
				}, nil)
				oSvc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantDone: false,
		},
		{
			name:  "bService.Update error returns false",
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
			wantDone: false,
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

			poller := NewAccrualPoller(mockClient, mockOSvc, mockBSvc)
			got := poller.handleOrder(context.Background(), tt.order)

			assert.Equal(t, tt.wantDone, got)
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
	)

	poller.Add("12345", "NEW", "user-uuid")

	assert.Len(t, poller.orders, 1)
	assert.Equal(t, "12345", poller.orders[0].number)
	assert.Equal(t, "NEW", poller.orders[0].status)
	assert.Equal(t, "user-uuid", poller.orders[0].userUUID)
}
