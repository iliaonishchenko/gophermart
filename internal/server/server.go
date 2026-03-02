package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/iliaonishchenko/gophermart/internal/logger"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/pkg/api"
	"go.uber.org/zap"
)

type Auth interface {
	Authenticate(ctx context.Context, login, password string) (string, error)
	GeneratePasswordHash(password string) (string, error)
	GenerateToken(login string) (string, error)
}

type UserService interface {
	Create(ctx context.Context, user *models.User) error
}

type Server struct {
	authService Auth
	userService UserService
}

func NewServer(authService Auth, userService UserService) *Server {
	return &Server{
		authService: authService,
		userService: userService,
	}
}

func (s *Server) PostAPIUserRegister(ctx context.Context, request api.PostAPIUserRegisterRequestObject) (api.PostAPIUserRegisterResponseObject, error) {
	if request.Body.Login == "" || request.Body.Password == "" {
		logger.Log.Debug("empty login or password")
		return api.PostAPIUserRegister400Response{}, nil
	}
	passwordHash, err := s.authService.GeneratePasswordHash(request.Body.Password)
	if err != nil {
		logger.Log.Error("could not hash password", logger.Err(err))
		return nil, fmt.Errorf("could not hash password: %w", err)
	}
	userToRegister := models.User{
		Login:        request.Body.Login,
		PasswordHash: passwordHash,
	}

	err = s.userService.Create(ctx, &userToRegister)
	if errors.Is(err, models.ErrUserAlreadyExists) {
		logger.Log.Info("user already exists", zap.String("login", request.Body.Login))
		return api.PostAPIUserRegister409Response{}, nil
	}
	if err != nil {
		logger.Log.Error("could not create user", logger.Err(err))
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	token, err := s.authService.GenerateToken(request.Body.Login)
	if err != nil {
		logger.Log.Error("could not generate token", logger.Err(err))
		return nil, fmt.Errorf("could not generate token: %w", err)
	}

	return api.PostAPIUserRegister200Response{
		Headers: api.PostAPIUserRegister200ResponseHeaders{
			Authorization: "Bearer " + token,
		},
	}, nil
}

func (s *Server) PostAPIUserLogin(ctx context.Context, request api.PostAPIUserLoginRequestObject) (api.PostAPIUserLoginResponseObject, error) {
	if request.Body.Login == "" || request.Body.Password == "" {
		logger.Log.Debug("empty login or password")
		return api.PostAPIUserLogin400Response{}, nil
	}

	token, err := s.authService.Authenticate(ctx, request.Body.Login, request.Body.Password)
	if errors.Is(err, models.ErrInvalidCredentials) {
		logger.Log.Error("invalid credentials", logger.Err(err))
		return api.PostAPIUserLogin401Response{}, nil
	}
	if err != nil {
		logger.Log.Error("could not authenticate", logger.Err(err))
		return api.PostAPIUserLogin500Response{}, nil
	}

	return api.PostAPIUserLogin200Response{
		Headers: api.PostAPIUserLogin200ResponseHeaders{
			Authorization: "Bearer " + token,
		},
	}, nil
}
