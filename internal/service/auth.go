package service

import (
	"context"
	"fmt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"io"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	emailScope    = "https://www.googleapis.com/auth/userinfo.email"
	profileScope  = "https://www.googleapis.com/auth/userinfo.profile"
	userInfo = "https://www.googleapis.com/oauth2/v2/userinfo"
)

type userProvider interface {
	CreateUser(ctx context.Context)
}

type token struct {
	t   *oauth2.Token
	cfg oauth2.Config
}

type Auth struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

func New(clientID, clientSecret, redirectURI string) Auth {
	return Auth{clientID: clientID, clientSecret: clientSecret, redirectURI: redirectURI}
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
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%w: %s", err, op)
	}
	fmt.Println(t)

	t.getUserInfo(ctx)

	// TODO: create user, save tokens, gen tokens
	return jwt.AccessToken{}, jwt.RefreshToken{}, nil
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

func (t token) getUserInfo(ctx context.Context) error {
	const op = "service.Auth.getUserInfo"

	res, err := t.cfg.Client(ctx, t.t).Get(userInfo)
	if err != nil {
		return fmt.Errorf("%w: %s", err, op)
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	fmt.Println(res)
	fmt.Println(string(b), "res bosy")

	return nil
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
