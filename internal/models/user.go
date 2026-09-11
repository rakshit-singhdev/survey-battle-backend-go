package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserRole string

const (
	UserRoleHost   UserRole = "host"
	UserRoleGuest  UserRole = "guest"
	UserRolePlayer UserRole = "player"
)

type Profile struct {
	FullName string  `bson:"full_name" json:"full_name"`
	Avatar   *string `bson:"avatar" json:"avatar"`
}

type EmailDetail struct {
	Email      string `bson:"email,omitempty" json:"email,omitempty"`
	IsVerified bool   `bson:"isVerified" json:"isVerified"`
}

type Auth struct {
	RefreshToken                  *string    `bson:"refreshToken" json:"-"`
	RefreshTokenExpiresAt         *time.Time `bson:"refreshTokenExpiresAt" json:"-"`
	PreviousRefreshToken          *string    `bson:"previousRefreshToken" json:"-"`
	PreviousRefreshTokenExpiresAt *time.Time `bson:"previousRefreshTokenExpiresAt" json:"-"`
	PasswordChangedAt             *time.Time `bson:"passwordChangedAt" json:"-"`
	PasswordResetToken            *string    `bson:"passwordResetToken" json:"-"`
	PasswordResetExpires          *time.Time `bson:"passwordResetExpires" json:"-"`
}

type User struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Profile     Profile       `bson:"profile" json:"profile"`
	EmailDetail EmailDetail   `bson:"emailDetail" json:"emailDetail"`
	Role        string        `bson:"role" json:"role"`
	Password    string        `bson:"password" json:"-"`
	Auth        Auth          `bson:"auth" json:"-"`
}
