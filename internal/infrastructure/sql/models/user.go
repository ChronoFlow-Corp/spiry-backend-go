package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID      `db:"user_id"`
	Email              string         `db:"user_email"`
	AvatarURL          string         `db:"user_avatar_url"`
	Name               string         `db:"user_name"`
	LastName           sql.NullString `db:"user_last_name"`
	Theme              string         `db:"user_theme"`
	GoogleAccessToken  string         `db:"user_google_access_token"`
	GoogleRefreshToken sql.NullString `db:"user_google_refresh_token"`
	CreatedAt          time.Time      `db:"user_created_at"`
	UpdatedAt          time.Time      `db:"user_updated_at"`
}

type UnloggedUser struct {
	ID        uuid.UUID `db:"unlogged_user_id"`
	IP        string    `db:"unlogged_user_ip"`
	PlanID    uuid.UUID `db:"unlogged_user_plan_id"`
	CreatedAt time.Time `db:"unlogged_user_created_at"`
	UpdatedAt time.Time `db:"unlogged_user_updated_at"`
}

type Plan struct {
	ID              uuid.UUID       `db:"plan_id"`
	Name            string          `db:"plan_name"`
	ModalitiesQuote modalitiesQuote `db:"plan_modalities_quote"`
	SubscriptionID  uuid.UUID       `db:"plan_subscription_id"`
	Level           uint            `db:"plan_level"`
	UserID          *uuid.UUID      `db:"plan_user_id"`
	CreatedAt       time.Time       `db:"plan_created_at"`
	UpdatedAt       time.Time       `db:"plan_updated_at"`
}

type Session struct {
	ID             uuid.UUID  `db:"session_id"`
	Token          string     `db:"session_token"`
	ExpiresAt      time.Time  `db:"session_expires_at"`
	LastLogin      time.Time  `db:"session_last_login"`
	Device         string     `db:"session_device"`
	UserID         *uuid.UUID `db:"session_user_id"`
	UnloggedUserID *uuid.UUID `db:"session_unlogged_user_id"`
	CreatedAt      time.Time  `db:"session_created_at"`
	UpdatedAt      time.Time  `db:"session_updated_at"`
}

type Subscription struct {
	ID              uuid.UUID        `db:"subscription_id"`
	Name            string           `db:"subscription_name"`
	ModalitiesQuote *modalitiesQuote `db:"subscription_modalities_quote"`
	Period          string           `db:"subscription_period"`
	Price           sql.NullString   `db:"subscription_price"`
	Level           int              `db:"subscription_level"`
	CreatedAt       time.Time        `db:"subscription_created_at"`
	UpdatedAt       time.Time        `db:"subscription_updated_at"`
}

type Model struct {
	ID         uuid.UUID `db:"model_id"`
	Name       string    `db:"model_name"`
	Modalities []string  `db:"model_modalities"`
	MinLevel   uint      `db:"model_min_level"`
	CreatedAt  time.Time `db:"model_created_at"`
	UpdatedAt  time.Time `db:"model_updated_at"`
}

type modalitiesQuote struct {
	MediaQuote       sql.NullInt64 `json:"media_quote"`
	TextContentQuote sql.NullInt64 `json:"text_content_quote"`
	ChattingQuote    sql.NullInt64 `json:"chatting_quote"`
}
