package repository

import "github.com/google/uuid"

type User struct {
	userID       uuid.UUID
	email        string
	accessToken  string
	refreshTokenGoogle string
	refreshToken string
}
