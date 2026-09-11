package service

import (
	"time"

	"survey-battle-backend-go/internal/config"
	"survey-battle-backend-go/internal/http/middleware"
	"survey-battle-backend-go/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	Config *config.JWTConfig
}

func (s *JWTService) GenerateAccessToken(user *models.User) (string, error) {
	now := time.Now()

	claims := middleware.Claims{
		ID:   user.ID.Hex(),
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.Config.AccessExpiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.Config.AccessSecret))
}
