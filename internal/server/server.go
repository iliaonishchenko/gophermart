package server

import (
	"context"
	"github.com/iliaonishchenko/gophermart/internal/models"
)

type Auth interface {
	Authenticate(ctx context.Context, login, password string) (string, error)
	GeneratePasswordHash(password string) (string, error)
	GenerateToken(uuid *string) (string, error)
}

type UserService interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
}

type OrderService interface {
	Create(ctx context.Context, order *models.Order) (*models.Order, error)
	Get(ctx context.Context, userUUID *string) ([]*models.Order, error)
}

type Server struct {
	authService  Auth
	userService  UserService
	orderService OrderService
}

func NewServer(authService Auth, userService UserService, orderService OrderService) *Server {
	return &Server{
		authService:  authService,
		userService:  userService,
		orderService: orderService,
	}
}
