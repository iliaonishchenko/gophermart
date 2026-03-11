package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/iliaonishchenko/gophermart/internal/config"
	"strings"
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
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(jwts.cfg.JwtExpire))),
		},
		UUID: *uuid,
	})
	tokenString, err := token.SignedString([]byte(jwts.cfg.JwtSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (jwts *JwtService) ParseToken(authHeader string) (string, error) {
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", fmt.Errorf("missing Bearer prefix")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwts.cfg.JwtSecret), nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	return claims.UUID, nil
}
