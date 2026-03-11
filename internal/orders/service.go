package orders

import (
	"context"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"go.uber.org/zap"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) (*models.Order, error)
	Get(ctx context.Context, uuid *string) ([]*models.Order, error)
	Update(ctx context.Context, order *models.Order) (*models.Order, error)
}

type Service struct {
	orderRepository OrderRepository
	log             *zap.Logger
}

func NewService(orderRepository OrderRepository, log *zap.Logger) *Service {
	return &Service{orderRepository: orderRepository, log: log}
}

func (s *Service) Create(ctx context.Context, order *models.Order) (*models.Order, error) {
	if !ValidateLuhn(order.Number) {
		return nil, models.ErrInvalidOrderFormat
	}
	return s.orderRepository.Create(ctx, order)
}

func (s *Service) Get(ctx context.Context, userUUID *string) ([]*models.Order, error) {
	return s.orderRepository.Get(ctx, userUUID)
}

func (s *Service) Update(ctx context.Context, order *models.Order) (*models.Order, error) {
	return s.orderRepository.Update(ctx, order)
}
