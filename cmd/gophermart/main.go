package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/iliaonishchenko/gophermart"
	"github.com/iliaonishchenko/gophermart/internal/accrual"
	"github.com/iliaonishchenko/gophermart/internal/auth"
	"github.com/iliaonishchenko/gophermart/internal/balance"
	"github.com/iliaonishchenko/gophermart/internal/config"
	"github.com/iliaonishchenko/gophermart/internal/logger"
	"github.com/iliaonishchenko/gophermart/internal/orders"
	"github.com/iliaonishchenko/gophermart/internal/server"
	"github.com/iliaonishchenko/gophermart/internal/users"
	"github.com/iliaonishchenko/gophermart/internal/withdrawals"
	"github.com/iliaonishchenko/gophermart/pkg/api"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.LoadConfiguration()
	if err != nil {
		return err
	}

	zlog, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		return err
	}
	err = gophermart.RunMigrations(db)
	if err != nil {
		return err
	}

	usersRepo := users.NewRepository(db, zlog)
	jwtService := auth.NewJwtService(cfg)
	authService := auth.NewAuthService(usersRepo, jwtService)
	userService := users.NewService(usersRepo)

	ordersRepo := orders.NewRepository(db, zlog)
	ordersService := orders.NewService(ordersRepo, zlog)

	withdrawalsRepo := withdrawals.NewRepository(db, zlog)
	withdrawalsService := withdrawals.NewService(withdrawalsRepo)

	balanceRepo := balance.NewRepository(db, zlog)
	balanceService := balance.NewService(balanceRepo)

	client := accrual.NewClient(cfg.AccrualAddr, zlog)
	accrualPoller := accrual.NewAccrualPoller(client, ordersService, balanceService, zlog)

	srv := server.NewServer(authService, userService, ordersService, withdrawalsService, balanceService, accrualPoller, zlog)

	r := chi.NewRouter()
	r.Use(logger.WithLogger(zlog))

	strictHandler := api.NewStrictHandler(srv, nil)

	r.Post("/api/user/register", strictHandler.PostAPIUserRegister)
	r.Post("/api/user/login", strictHandler.PostAPIUserLogin)

	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware(jwtService))
		r.Post("/api/user/orders", strictHandler.PostAPIUserOrders)
		r.Get("/api/user/orders", strictHandler.GetAPIUserOrders)
		r.Post("/api/user/balance/withdraw", strictHandler.PostAPIUserBalanceWithdraw)
		r.Get("/api/user/withdrawals", strictHandler.GetAPIUserWithdrawals)
		r.Get("/api/user/balance", strictHandler.GetAPIUserBalance)
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go accrualPoller.Run(ctx)

	errCh := make(chan error, 1)
	httpServer := &http.Server{Addr: cfg.ServerAddr, Handler: r}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	return httpServer.Shutdown(context.Background())
}
