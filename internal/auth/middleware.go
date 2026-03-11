package auth

import (
	"context"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"log"
	"net/http"
)

type contextKey string

const UserUUIDKey contextKey = "user_uuid"

func AuthMiddleware(jwtService *JwtService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			uuid, err := jwtService.ParseToken(authHeader)
			if err != nil {
				log.Println(err.Error())
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserUUIDKey, uuid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserUUID(ctx context.Context) (string, error) {
	val := ctx.Value(UserUUIDKey)
	if val == nil {
		return "", models.ErrNoUserInContext
	}
	uuid, ok := val.(string)
	if !ok {
		return "", models.ErrInvalidUserType
	}
	return uuid, nil
}
