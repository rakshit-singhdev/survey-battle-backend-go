package repository

import (
	"context"
	"errors"

	"survey-battle-backend-go/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)

type UserRepository struct {
	Collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		Collection: db.Collection("users"),
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	err := r.Collection.FindOne(ctx, bson.M{"emailDetail.email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.Collection.InsertOne(ctx, user)

	return err
}
