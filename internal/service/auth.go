package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	emailScope   = "https://www.googleapis.com/auth/userinfo.email"
	profileScope = "https://www.googleapis.com/auth/userinfo.profile"
	userInfoURL  = "https://www.googleapis.com/oauth2/v3/userinfo"
)

type token struct {
	t   *oauth2.Token
	cfg oauth2.Config
}

type userInfo struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type userProvider interface {
	SaveUser(ctx context.Context, u repository.User) error
}

type Auth struct {
	clientID     string
	clientSecret string
	redirectURI  string
	userProvider userProvider
	jwt          jwt.JWT
}

func New(clientID, clientSecret, redirectURI string, up userProvider, jwt jwt.JWT) Auth {
	return Auth{clientID: clientID, clientSecret: clientSecret, redirectURI: redirectURI, userProvider: up, jwt: jwt}
}

func (a Auth) GetAuthCodeURI() string {
	cfg := a.buildConfig(emailScope, profileScope)
	userID := uuid.New()

	return cfg.AuthCodeURL(fmt.Sprintf("userID=%s", userID), oauth2.AccessTypeOffline)
}

func (a Auth) Login(ctx context.Context, state map[string]string, code string) (jwt.AccessToken, jwt.RefreshToken, error) {
	const op = "service.Auth.Login"

	t, err := a.exchangeCode(ctx, code)
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	info, err := t.getUserInfo(ctx)
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := uuid.Parse(state["userID"])
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	access, refresh, err := a.jwt.NewPair(userID.String())
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	u := repository.NewUser(userID, info.Email, t.t.AccessToken, t.t.RefreshToken, refresh.Raw)

	err = a.userProvider.SaveUser(ctx, u)
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	return access, refresh, nil
}

func (a Auth) exchangeCode(ctx context.Context, code string) (token, error) {
	const op = "service.Auth.exchangeCode"

	cfg := a.buildConfig(emailScope, profileScope)

	t, err := cfg.Exchange(ctx, code)
	if err != nil {
		return token{}, newAuthError(err, fmt.Sprintf("%s: %s", op, "cannot exchange code"))
	}

	return token{t: t, cfg: cfg}, nil
}

func (t token) getUserInfo(ctx context.Context) (userInfo, error) {
	const op = "service.Auth.getUserInfo"

	res, err := t.cfg.Client(ctx, t.t).Get(userInfoURL)
	if err != nil {
		return userInfo{}, newAuthError(err, fmt.Sprintf("%s: %s", op, "cannot get user info"))
	}
	defer res.Body.Close()

	var usInfo userInfo

	err = json.NewDecoder(res.Body).Decode(&usInfo)
	if err != nil {
		return userInfo{}, newAuthError(err, fmt.Sprintf("%s: %s", op, "cannot decode user info"))
	}

	return usInfo, nil
}

func (a Auth) buildConfig(scopes ...string) oauth2.Config {
	return oauth2.Config{
		ClientID:     a.clientID,
		ClientSecret: a.clientSecret,
		RedirectURL:  a.redirectURI,
		Endpoint:     google.Endpoint,
		Scopes:       scopes,
	}
}
