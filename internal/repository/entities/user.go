package entities

import (
	"time"

	"github.com/google/uuid"
)

type UserIDCtxKey struct{}

type User struct {
	ID                 uuid.UUID
	Email              string
	AccessTokenGoogle  string
	RefreshTokenGoogle string
	RefreshToken       string
	Language string
	Plan               Plan
	Admin              bool
	Theme string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewUser(
	id uuid.UUID,
	email, accessTokenGoogle, refreshTokenGoogle, refreshToken, language string,
	admin bool,
	plan Plan,
	theme string) User {
	return User{
		ID: id,
		Email: email,
		AccessTokenGoogle: accessTokenGoogle,
		RefreshTokenGoogle: refreshTokenGoogle,
		RefreshToken: refreshToken,
		Language: language,
		Admin: admin,
		Plan: plan,
		Theme: theme,
	}
}
