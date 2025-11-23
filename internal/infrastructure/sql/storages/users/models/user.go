package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID      `db:"id"`
	Email              string         `db:"email"`
	AvatarURL          string         `db:"avatar_url"`
	Name               string         `db:"name"`
	LastName           sql.NullString `db:"last_name"`
	Theme              string         `db:"theme"`
	GoogleAccessToken  string         `db:"google_access_token"`
	GoogleRefreshToken sql.NullString `db:"google_refresh_token"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
}
