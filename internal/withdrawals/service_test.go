package withdrawals

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/internal/withdrawals/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	defaultTime := time.Now()
	tests := []struct {
		name             string
		withdrawal       *models.Withdrawal
		resultWithdrawal *models.Withdrawal
		resultError      error
	}{
		{
			name:             "success with valid withdrawal",
			withdrawal:       &models.Withdrawal{Order: "12345678903", Sum: 500},
			resultWithdrawal: &models.Withdrawal{Order: "12345678903", Sum: 500, ProcessedAt: defaultTime},
			resultError:      nil,
		},
		{
			name:             "error with invalid order number",
			withdrawal:       &models.Withdrawal{Order: "12345678901", Sum: 500},
			resultWithdrawal: nil,
			resultError:      models.ErrWithdrawalNonExistentOrder,
		},
		{
			name:             "error from repository",
			withdrawal:       &models.Withdrawal{Order: "12345678903", Sum: 500},
			resultWithdrawal: nil,
			resultError:      models.ErrOrderExistsDifferentUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockWithdrawalsRepository(ctrl)
			if !errors.Is(tt.resultError, models.ErrWithdrawalNonExistentOrder) {
				mockRepo.EXPECT().
					Create(gomock.Any(), tt.withdrawal).
					Return(tt.resultWithdrawal, tt.resultError)
			}

			svc := NewService(mockRepo)

			order, err := svc.Create(context.Background(), tt.withdrawal)

			assert.Equal(t, tt.resultError, err)
			assert.Equal(t, tt.resultWithdrawal, order)
		})
	}
}

func TestGet(t *testing.T) {
	defaultTime := time.Now()
	errorFromRepo := errors.New("error from repository")
	tests := []struct {
		name              string
		resultWithdrawals []*models.Withdrawal
		resultError       error
	}{
		{
			name: "success returns orders",
			resultWithdrawals: []*models.Withdrawal{
				{Order: "12345678903", Sum: 500, ProcessedAt: defaultTime},
				{Order: "12345678905", Sum: 501, ProcessedAt: defaultTime},
			},
			resultError: nil,
		},
		{
			name:              "error returns error",
			resultWithdrawals: nil,
			resultError:       errorFromRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockWithdrawalsRepository(ctrl)
			mockRepo.EXPECT().
				Get(gomock.Any()).
				Return(tt.resultWithdrawals, tt.resultError)

			svc := NewService(mockRepo)

			orders, err := svc.Get(context.Background())

			assert.Equal(t, tt.resultError, err)
			assert.Equal(t, tt.resultWithdrawals, orders)
		})
	}
}
