package main

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/iliaonishchenko/gophermart"
	"github.com/iliaonishchenko/gophermart/internal/auth"
	"github.com/iliaonishchenko/gophermart/internal/config"
	"github.com/iliaonishchenko/gophermart/internal/logger"
	"github.com/iliaonishchenko/gophermart/internal/orders"
	"github.com/iliaonishchenko/gophermart/internal/server"
	"github.com/iliaonishchenko/gophermart/internal/users"
	"github.com/iliaonishchenko/gophermart/pkg/api"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"
)

func main() {

	_ = context.Background()
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

	srv := server.NewServer(authService, userService, ordersService)

	r := chi.NewRouter()
	r.Use(logger.WithLogger)

	strictHandler := api.NewStrictHandler(srv, nil)

	r.Post("/api/user/register", strictHandler.PostAPIUserRegister)
	r.Post("/api/user/login", strictHandler.PostAPIUserLogin)

	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware(jwtService))
		r.Post("/api/user/orders", strictHandler.PostAPIUserOrders)
		r.Get("/api/user/orders", strictHandler.GetAPIUserOrders)
	})

	handler := r

	err = http.ListenAndServe(cfg.ServerAddr, handler)
	if err != nil {
		return
	}
}
