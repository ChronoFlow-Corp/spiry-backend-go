package connect

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/model"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	emailScope   = "https://www.googleapis.com/auth/userinfo.email"
	profileScope = "https://www.googleapis.com/auth/userinfo.profile"
)

const userInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"

type GoogleOauth struct {
	config *oauth2.Config
}

func NewGoogleOauth(clientID, clientSecret, redirectURI string) *GoogleOauth {
	return &GoogleOauth{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				emailScope,
				profileScope,
			},
		},
	}
}

func NewGoogleOauthFx(cfg *config.Config) *GoogleOauth {
	return NewGoogleOauth(
		cfg.GoogleAuth.ClientID,
		cfg.GoogleAuth.ClientSecret,
		cfg.GoogleAuth.RedirectURI,
	)
}

func (g *GoogleOauth) GetAuthCodeURI(device string) string {
	stateDevice := "device" + "=" + device

	authURL := g.config.AuthCodeURL(
		stateDevice,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	return authURL
}

func (g *GoogleOauth) ExchangeCode(ctx context.Context, code string) (model.OauthInfo, error) {
	const op = "infrastructure.auth.connect.google.Exchange"

	t, err := g.config.Exchange(ctx, code)
	if err != nil {
		return model.OauthInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	rs, err := g.config.Client(ctx, t).Get(userInfoURL)
	if err != nil {
		return model.OauthInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	var u userInfo

	err = json.NewDecoder(rs.Body).Decode(&u)
	if err != nil {
		return model.OauthInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	return model.OauthInfo{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresAt:    t.Expiry,
		Email:        u.Email,
		Name:         u.Name,
		AvatarURL:    u.Picture,
		LastName:     u.FamilyName,
	}, nil
}

type userInfo struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
}
