package repository

import "github.com/google/uuid"

type User struct {
	userID             uuid.UUID
	email              string
	accessTokenGoogle  string
	refreshTokenGoogle string
	refreshToken       string
}

func NewUser(userID uuid.UUID, email, accessTokenGoogle, refreshTokenGoogle, refreshToken string) User {
	return User{
		userID:             userID,
		email:              email,
		accessTokenGoogle:  accessTokenGoogle,
		refreshTokenGoogle: refreshTokenGoogle,
		refreshToken:       refreshToken,
	}
}

func (u User) ID() uuid.UUID {
	return u.userID
}

func (u User) Email() string {
	return u.email
}

func (u User) AccessTokenGoogle() string {
	return u.accessTokenGoogle
}

func (u User) RefreshTokenGoogle() string {
	return u.refreshTokenGoogle
}

func (u User) RefreshToken() string {
	return u.refreshToken
}
