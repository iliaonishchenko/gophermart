package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/iliaonishchenko/gophermart/internal/config"
	"time"
)

type JwtService struct {
	cfg *config.Config
}

type Claims struct {
	jwt.RegisteredClaims
	UUID string
}

func NewJwtService(cfg *config.Config) *JwtService {
	return &JwtService{cfg: cfg}
}

func (jwts *JwtService) GenerateToken(uuid *string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwts.cfg.JwtExpire))),
		},
		UUID: *uuid,
	})
	tokenString, err := token.SignedString([]byte(jwts.cfg.JwtSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
