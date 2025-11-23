package model

import "time"

type OauthInfo struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Email        string
	AvatarURL    string
	Name         string
	LastName     string
}
