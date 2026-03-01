package users

import (
	"context"
	"github.com/iliaonishchenko/gophermart/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
}

type Service struct {
	userRepository UserRepository
}

func NewService(userRepository UserRepository) *Service {
	return &Service{userRepository: userRepository}
}

func (s *Service) Create(ctx context.Context, user *models.User) error {
	return s.userRepository.Create(ctx, user)
}
