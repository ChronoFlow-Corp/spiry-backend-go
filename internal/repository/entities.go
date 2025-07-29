package repository

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	userID             uuid.UUID
	email              string
	accessTokenGoogle  string
	refreshTokenGoogle string
	refreshToken       string
	CreatedAt time.Time
	UpdatedAt time.Time
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

type Chat struct{
	ID uuid.UUID
	Title string
	UserID uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages []Message
}

func NewChat(id, userID uuid.UUID, title string) Chat {
	return Chat{
		ID: id,
		UserID:  userID,
		Title:   title,
	}
}

func (c Chat) GetID() uuid.UUID {
	return c.UserID
}

func (c Chat) GetTitle() string {
	return c.Title
}

func (c Chat) GetUserID() uuid.UUID {
	return c.UserID
}

type Message struct {
	ID uuid.UUID
	Answer string
	Question string
	CreatedAt time.Time
	ChatID uuid.UUID
	UserID uuid.UUID
}

func NewMessage(id uuid.UUID, question, answer string, userID uuid.UUID) Message {
	return Message{
		ID: id,
		Question: question,
		Answer: answer,
		UserID: userID,
	}
}

func (m Message) GetID() uuid.UUID {
	return m.ID
}

func (m Message) GetQuestion() string {
	return m.Question
}

func (m Message) GetUserID() uuid.UUID {
	return m.UserID
}

func (m Message) GetAnswer() string {
	return m.Answer
}

func (m Message) GetCreatedAt() time.Time {
	return m.CreatedAt
}