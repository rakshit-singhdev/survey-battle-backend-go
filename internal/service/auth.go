package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"survey-battle-backend-go/internal/models"
	"survey-battle-backend-go/internal/repository"
	"survey-battle-backend-go/internal/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthService struct {
	UserRepository *repository.UserRepository
	JWTService     *JWTService
}

func (s *AuthService) Register(
	ctx context.Context,
	fullName string,
	email string,
	password string,
) (*models.User, string, string, error) {

	// Check whether email already exists.
	_, err := s.UserRepository.FindByEmail(ctx, email)

	if err == nil {
		return nil, "", "", repository.ErrEmailExists
	}

	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, "", "", err
	}

	// Hash password using Argon2id.
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", "", err
	}

	user := &models.User{
		Profile: models.Profile{
			FullName: fullName,
		},

		EmailDetail: models.EmailDetail{
			Email:      email,
			IsVerified: false,
		},

		Role:     string(models.UserRoleHost),
		Password: hashedPassword,

		Auth: models.Auth{},
	}

	// Generate MongoDB ID.
	// The Mongo driver can also generate this during InsertOne.
	// Keeping it here makes the response immediately available.
	user.ID = newObjectID()

	// Generate tokens.
	accessToken, err := s.JWTService.GenerateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, "", "", err
	}

	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)

	user.Auth.RefreshToken = &refreshToken
	user.Auth.RefreshTokenExpiresAt = &refreshExpiry

	// Save user.
	if err := s.UserRepository.Create(ctx, user); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (*models.User, string, string, error) {

	user, err := s.UserRepository.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", "", err
	}

	// Verify Argon2id password.
	if !utils.ComparePassword(password, user.Password) {
		return nil, "", "", errors.New("invalid credentials")
	}

	accessToken, err := s.JWTService.GenerateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, "", "", err
	}

	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)

	user.Auth.RefreshToken = &refreshToken
	user.Auth.RefreshTokenExpiresAt = &refreshExpiry

	// You would update these fields in MongoDB here.
	//
	// Example:
	//
	// err = s.UserRepository.UpdateRefreshToken(
	//     ctx,
	//     user.ID,
	//     refreshToken,
	//     refreshExpiry,
	// )

	return user, accessToken, refreshToken, nil
}

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 64)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func newObjectID() bson.ObjectID {
	return bson.NewObjectID()
}
