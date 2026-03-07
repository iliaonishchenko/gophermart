package main

import (
	"context"
	"database/sql"
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

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	parseFlags(cfg)

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		log.Fatal(err)
	}
	err = gophermart.RunMigrations(db)
	if err != nil {
		log.Fatal(err)
	}

	usersRepo := users.NewRepository(db)
	jwtService := auth.NewJwtService(cfg)
	authService := auth.NewAuthService(usersRepo, jwtService)
	userService := users.NewService(usersRepo)

	ordersRepo := orders.NewRepository(db)
	ordersService := orders.NewService(ordersRepo)

	withdrawalsRepo := withdrawals.NewRepository(db)
	withdrawalsService := withdrawals.NewService(withdrawalsRepo)

	balanceRepo := balance.NewRepository(db)
	balanceService := balance.NewService(balanceRepo)

	client := accrual.NewClient(cfg.AccrualAddr)
	accrualPoller := accrual.NewAccrualPoller(client, ordersService, balanceService)

	srv := server.NewServer(authService, userService, ordersService, withdrawalsService, balanceService, accrualPoller)

	r := chi.NewRouter()
	r.Use(logger.WithLogger)

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

	httpServer := &http.Server{Addr: cfg.ServerAddr, Handler: r}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	httpServer.Shutdown(context.Background())
}
