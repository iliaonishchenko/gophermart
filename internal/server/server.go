package server

import (
	"context"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"go.uber.org/zap"
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

type WithdrawalService interface {
	Create(ctx context.Context, withdrawalToCreate *models.Withdrawal) (*models.Withdrawal, error)
	Get(ctx context.Context) ([]*models.Withdrawal, error)
}

type BalanceService interface {
	Get(ctx context.Context) (*models.Balance, error)
}

type AccrualService interface {
	Run(ctx context.Context)
	Add(number string, status string, userUUID string)
}

type Server struct {
	authService       Auth
	userService       UserService
	orderService      OrderService
	withdrawalService WithdrawalService
	balanceService    BalanceService
	accrualService    AccrualService
	log               *zap.Logger
}

func NewServer(auth Auth, userService UserService, orderService OrderService, withdrawalService WithdrawalService, balanceService BalanceService, accrualService AccrualService, log *zap.Logger) *Server {
	return &Server{
		authService:       auth,
		userService:       userService,
		orderService:      orderService,
		withdrawalService: withdrawalService,
		balanceService:    balanceService,
		accrualService:    accrualService,
		log:               log,
	}
}
