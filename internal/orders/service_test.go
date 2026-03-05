package orders

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/internal/orders/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	userUUID := uuid.New().String()
	defaultTime := time.Now()
	tests := []struct {
		name        string
		order       *models.Order
		resultOrder *models.Order
		resultError error
	}{
		{
			name:        "success with valid order",
			order:       &models.Order{Number: "12345678903", UserID: userUUID, Status: models.OrderStatusNew, Accrual: nil},
			resultOrder: &models.Order{Number: "12345678903", UserID: userUUID, Status: models.OrderStatusNew, Accrual: nil, UploadedAt: defaultTime},
			resultError: nil,
		},
		{
			name:        "error with invalid order number",
			order:       &models.Order{Number: "12345678901", UserID: userUUID, Status: models.OrderStatusNew, Accrual: nil},
			resultOrder: nil,
			resultError: models.ErrInvalidOrderFormat,
		},
		{
			name:        "error from repository",
			order:       &models.Order{Number: "12345678903", UserID: userUUID, Status: models.OrderStatusNew, Accrual: nil},
			resultOrder: nil,
			resultError: models.ErrOrderExistsDifferentUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockOrderRepository(ctrl)
			if !errors.Is(tt.resultError, models.ErrInvalidOrderFormat) {
				mockRepo.EXPECT().
					Create(gomock.Any(), tt.order).
					Return(tt.resultOrder, tt.resultError)
			}

			svc := NewService(mockRepo)

			order, err := svc.Create(context.Background(), tt.order)

			assert.Equal(t, tt.resultError, err)
			assert.Equal(t, tt.resultOrder, order)
		})
	}
}

func TestGet(t *testing.T) {
	userUUID := uuid.New().String()
	defaultTime := time.Now()
	errorFromRepo := errors.New("error from repository")
	tests := []struct {
		name         string
		resultOrders []*models.Order
		resultError  error
	}{
		{
			name: "success returns orders",
			resultOrders: []*models.Order{
				{Number: "12345678903", UserID: userUUID, Status: models.OrderStatusNew, Accrual: nil, UploadedAt: defaultTime},
				{Number: "12345678905", UserID: userUUID, Status: models.OrderStatusNew, Accrual: nil, UploadedAt: defaultTime},
			},
			resultError: nil,
		},
		{
			name:         "error returns error",
			resultOrders: nil,
			resultError:  errorFromRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockOrderRepository(ctrl)
			mockRepo.EXPECT().
				Get(gomock.Any(), &userUUID).
				Return(tt.resultOrders, tt.resultError)

			svc := NewService(mockRepo)

			orders, err := svc.Get(context.Background(), &userUUID)

			assert.Equal(t, tt.resultError, err)
			assert.Equal(t, tt.resultOrders, orders)
		})
	}
}
