package auth

import (
	"context"
	"fmt"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type JwtGenerator interface {
	GenerateToken(login string) (string, error)
}

type UsersRepository interface {
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

type AuthService struct {
	userRepo     UsersRepository
	jwtGenerator JwtGenerator
}

func NewAuthService(userRepo UsersRepository, jwtGen JwtGenerator) *AuthService {
	return &AuthService{userRepo: userRepo, jwtGenerator: jwtGen}
}

func (as *AuthService) Authenticate(ctx context.Context, login, password string) (string, error) {
	user, err := as.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", fmt.Errorf("could not get user by login: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", models.ErrInvalidCredentials
	}
	token, err := as.jwtGenerator.GenerateToken(user.Login)
	if err != nil {
		return "", fmt.Errorf("could not generate token: %w", err)
	}
	return token, nil
}

func (as *AuthService) GenerateToken(login string) (string, error) {
	return as.jwtGenerator.GenerateToken(login)
}

func (as *AuthService) GeneratePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("could not hash password: %w", err)
	}
	return string(hash), nil
}
