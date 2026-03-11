package balance

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/iliaonishchenko/gophermart/internal/balance/mocks"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	errorFromRepo := errors.New("error from repository")
	tests := []struct {
		name          string
		resultBalance *models.Balance
		resultError   error
	}{
		{
			name:          "success returns balance",
			resultBalance: &models.Balance{Current: 500.5, Withdrawn: 42},
			resultError:   nil,
		},
		{
			name:          "success with zero balance",
			resultBalance: &models.Balance{Current: 0, Withdrawn: 0},
			resultError:   nil,
		},
		{
			name:          "error returns error",
			resultBalance: nil,
			resultError:   errorFromRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBalanceRepository(ctrl)
			mockRepo.EXPECT().
				Get(gomock.Any()).
				Return(tt.resultBalance, tt.resultError)

			svc := NewService(mockRepo)

			balance, err := svc.Get(context.Background())

			assert.Equal(t, tt.resultError, err)
			assert.Equal(t, tt.resultBalance, balance)
		})
	}
}
