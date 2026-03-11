package withdrawals

import (
	"context"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/internal/orders"
)

type WithdrawalsRepository interface {
	Create(ctx context.Context, withdrawalToCreate *models.Withdrawal) (*models.Withdrawal, error)
	Get(ctx context.Context) ([]*models.Withdrawal, error)
}

type Service struct {
	repo WithdrawalsRepository
}

func NewService(repo WithdrawalsRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, withdrawalToCreate *models.Withdrawal) (*models.Withdrawal, error) {
	ok := orders.ValidateLuhn(withdrawalToCreate.Order)
	if !ok {
		return nil, models.ErrWithdrawalNonExistentOrder
	}
	return s.repo.Create(ctx, withdrawalToCreate)
}

func (s *Service) Get(ctx context.Context) ([]*models.Withdrawal, error) {
	return s.repo.Get(ctx)
}
