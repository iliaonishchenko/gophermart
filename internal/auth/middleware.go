package auth

import (
	"context"
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

func GetUserUUID(ctx context.Context) (string, bool) {
	uuid, ok := ctx.Value(UserUUIDKey).(string)
	return uuid, ok
}
