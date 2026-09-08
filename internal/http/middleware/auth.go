package middleware

import (
	"context"
	"net/http"
	"strings"

	"survey-battle-backend-go/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type User struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type contextKey string

const UserContextKey contextKey = "user" //using custom key types prevents accidental collisions with keys from other packages

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Extract the token from the Authorization header
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(
					w,
					"Missing Authorization header",
					http.StatusUnauthorized,
				)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)

			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(
					w,
					"Invalid Authorization header format",
					http.StatusUnauthorized,
				)
				return
			}

			tokenString := parts[1]

			claims := &Claims{}

			token, err := jwt.ParseWithClaims(
				tokenString,
				claims,
				func(token *jwt.Token) (any, error) {
					return []byte(cfg.JWT.AccessSecret), nil
				},
			)

			if err != nil || !token.Valid {
				http.Error(
					w,
					"Unauthorized",
					http.StatusUnauthorized,
				)
				return
			}

			user := User{
				ID:   claims.ID,
				Role: claims.Role,
			}

			ctx := context.WithValue(
				r.Context(),
				UserContextKey,
				user,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}