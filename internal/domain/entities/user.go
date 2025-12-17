package entities

import (
	"regexp"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

var emailRegex = regexp.MustCompile(`(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

const (
	maxLenEmail  = 256
	maxLenNames  = 256
	maxLenTokens = 512
	maxLenAvatar = 512
)

type Theme string

const (
	ThemeLight Theme = "light"
	ThemeDark  Theme = "dark"
	ThemeAuto  Theme = "auto"
)

// User is entity of user.
type User struct {
	ID                 uuid.UUID
	Email              string
	AvatarURL          string
	Name               string
	LastName           string
	Theme              Theme
	GoogleAccessToken  string
	GoogleRefreshToken string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewUser(
	email string,
	firstName string,
	lastName string,
	avatarURL string,
	theme Theme,
	googleAccessToken string,
	googleRefreshToken string,
) *User {
	return &User{
		ID:                 uuid.New(),
		Email:              email,
		AvatarURL:          avatarURL,
		Name:               firstName,
		LastName:           lastName,
		Theme:              theme,
		GoogleAccessToken:  googleAccessToken,
		GoogleRefreshToken: googleRefreshToken,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now().Add(time.Millisecond),
	}
}

func (u *User) Validate() error {
	if len(u.Email) > maxLenEmail {
		return domain.NewValidationError(nil, "email", "email is too long")
	}

	if u.Email == "" {
		return domain.NewValidationError(nil, "email", "email is required")
	}

	if !emailRegex.MatchString(u.Email) {
		return domain.NewValidationError(nil, "email", "email is invalid")
	}

	if len(u.Name) > maxLenNames {
		return domain.NewValidationError(nil, "name", "name too long")
	}

	if len(u.LastName) > maxLenTokens {
		return domain.NewValidationError(nil, "lastName", "lastName too long")
	}

	if len(u.AvatarURL) > maxLenAvatar {
		return domain.NewValidationError(nil, "avatarURL", "avatarURL is too long")
	}

	if len(u.GoogleAccessToken) > maxLenTokens {
		return domain.NewValidationError(
			nil,
			"googleAccessToken",
			"googleAccessToken too long",
		)
	}

	if len(u.GoogleRefreshToken) > maxLenTokens {
		return domain.NewValidationError(
			nil,
			"googleRefreshToken",
			"googleRefreshToken too long",
		)
	}

	if u.Theme != ThemeAuto && u.Theme != ThemeDark && u.Theme != ThemeLight {
		return domain.NewValidationError(
			nil,
			"theme",
			"theme must be one of [light, dark, auto]",
		)
	}

	if u.CreatedAt.After(u.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
