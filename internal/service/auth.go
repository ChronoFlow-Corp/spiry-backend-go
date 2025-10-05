package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
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
	SaveUser(ctx context.Context, u entities.User) error
	UpdateUser(ctx context.Context, u entities.User) error
	UpdateRefreshToken(ctx context.Context, refreshToken string) error
	GetUserByID(ctx context.Context) (entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (entities.User, error)
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

	return cfg.AuthCodeURL("", oauth2.AccessTypeOffline)
}

func (a Auth) Refresh(ctx context.Context, token string) (jwt.AccessToken, jwt.RefreshToken, error) {
	const op = "repository.auth.Refresh"

	t, err := a.jwt.ParseRefresh(token)
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := uuid.Parse(t.Claims.Issuer)
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	ctx = context.WithValue(ctx, entities.UserIDCtxKey{}, userID)

	u, err := a.userProvider.GetUserByID(ctx)
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	if u.RefreshToken != t.Raw {
		return jwt.AccessToken{}, jwt.RefreshToken{},
			fmt.Errorf("%s: %w", op, errors.New("invalid refresh token"))
	}

	access, refresh, err := a.jwt.NewPair(userID.String())
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}
	err = a.userProvider.UpdateRefreshToken(ctx, refresh.Raw)
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	return access, refresh, nil
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

	user, err := a.userProvider.GetUserByEmail(ctx, info.Email)
	if err != nil {
		var notFoundErr *repository.ErrorNotFound
		if errors.As(err, &notFoundErr) {
			userID := uuid.New()
			access, refresh, err := a.jwt.NewPair(userID.String())

			if err != nil {
				return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
			}

			u := entities.NewUser(
				userID,
				info.Email,
				info.Name,
				info.Picture,
				t.t.AccessToken,
				t.t.RefreshToken,
				refresh.Raw,
				"en",
				false,
				entities.NewFreePlan(),
				"light",
			)

			err = a.registerUser(ctx, u)
			if err != nil {
				return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
			}

			return access, refresh, nil
		}

		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	access, refresh, err := a.jwt.NewPair(user.ID.String())
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	user.RefreshToken = refresh.Raw
	user.AccessTokenGoogle = t.t.AccessToken
	user.RefreshTokenGoogle = t.t.RefreshToken

	err = a.updateUser(ctx, user)
	if err != nil {
		return jwt.AccessToken{}, jwt.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	return access, refresh, nil
}

func (a Auth) registerUser(ctx context.Context, u entities.User) error {
	const op = "service.Auth.registerUser"

	err := a.userProvider.SaveUser(ctx, u)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a Auth) updateUser(ctx context.Context, u entities.User) error {
	const op = "service.Auth.updateUser"

	err := a.userProvider.UpdateUser(ctx, u)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a Auth) GetUserInfo(ctx context.Context) (entities.User, error) {
	const op = "service.Auth.GetUserInfo"

	u, err := a.userProvider.GetUserByID(ctx)
	if err != nil {
		return entities.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return u, nil
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
