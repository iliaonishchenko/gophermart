package balance

import (
	"context"
	"github.com/iliaonishchenko/gophermart/internal/models"
)

type BalanceRepository interface {
	Get(ctx context.Context) (*models.Balance, error)
}

type Service struct {
	repo BalanceRepository
}

func NewService(repo BalanceRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context) (*models.Balance, error) {
	return s.repo.Get(ctx)
}
