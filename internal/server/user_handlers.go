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

func (s *Server) PostAPIUserRegister(ctx context.Context, request api.PostAPIUserRegisterRequestObject) (api.PostAPIUserRegisterResponseObject, error) {
	if request.Body.Login == "" || request.Body.Password == "" {
		s.log.Debug("empty login or password")
		return api.PostAPIUserRegister400Response{}, nil
	}
	passwordHash, err := s.authService.GeneratePasswordHash(request.Body.Password)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %w", err)
	}
	userToRegister := models.User{
		Login:        request.Body.Login,
		PasswordHash: passwordHash,
	}

	user, err := s.userService.Create(ctx, &userToRegister)
	if errors.Is(err, models.ErrUserAlreadyExists) {
		s.log.Info("user already exists", zap.String("login", request.Body.Login))
		return api.PostAPIUserRegister409Response{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	token, err := s.authService.GenerateToken(user.ID)
	if err != nil {
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
		s.log.Debug("empty login or password")
		return api.PostAPIUserLogin400Response{}, nil
	}

	token, err := s.authService.Authenticate(ctx, request.Body.Login, request.Body.Password)
	if errors.Is(err, models.ErrInvalidCredentials) {
		s.log.Error("invalid credentials", logger.Err(err))
		return api.PostAPIUserLogin401Response{}, nil
	}
	if err != nil {
		s.log.Error("could not authenticate", logger.Err(err))
		return api.PostAPIUserLogin500Response{}, nil
	}

	return api.PostAPIUserLogin200Response{
		Headers: api.PostAPIUserLogin200ResponseHeaders{
			Authorization: "Bearer " + token,
		},
	}, nil
}