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
	userInfoURL  = "https://www.googleapis.com/oauth2/v2/userinfo"
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

	return cfg.AuthCodeURL(userID.String())
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

	if err := a.userProvider.SaveUser(ctx, u); err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	// TODO: create user, save tokens, gen tokens
	return access, refresh, nil
}

func (a Auth) exchangeCode(ctx context.Context, code string) (token, error) {
	const op = "service.Auth.exchangeCode"

	cfg := a.buildConfig(emailScope, profileScope)

	t, err := cfg.Exchange(ctx, code)
	if err != nil {
		return token{}, fmt.Errorf("cannot exchange code %s: %v", op, err)
	}

	return token{t: t, cfg: cfg}, nil
}

func (t token) getUserInfo(ctx context.Context) (userInfo, error) {
	const op = "service.Auth.getUserInfo"

	res, err := t.cfg.Client(ctx, t.t).Get(userInfoURL)
	if err != nil {
		return userInfo{}, fmt.Errorf("%s: %w", op, err)
	}
	defer res.Body.Close()

	var usInfo userInfo

	if err := json.NewDecoder(res.Body).Decode(&usInfo); err != nil {
		return userInfo{}, fmt.Errorf("%s: %w", op, err)
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
