package orders

import (
	"context"
	"github.com/iliaonishchenko/gophermart/internal/logger"
	"github.com/iliaonishchenko/gophermart/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) (*models.Order, error)
	Get(ctx context.Context, uuid *string) ([]*models.Order, error)
}

type Service struct {
	orderRepository OrderRepository
}

func NewService(orderRepository OrderRepository) *Service {
	return &Service{orderRepository: orderRepository}
}

func (s *Service) Create(ctx context.Context, order *models.Order) (*models.Order, error) {
	if !ValidateLuhn(order.Number) {
		logger.Log.Error("invalid order number")
		return nil, models.ErrInvalidOrderFormat
	}
	return s.orderRepository.Create(ctx, order)
}

func (s *Service) Get(ctx context.Context, userUUID *string) ([]*models.Order, error) {
	return s.orderRepository.Get(ctx, userUUID)
}
