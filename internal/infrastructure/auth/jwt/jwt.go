package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/model"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWT struct {
	accessSecretPrivate []byte
	accessSecretPublic  []byte
	refreshSecret       []byte
	accessExpires       time.Duration
	refreshExpires      time.Duration
}
type Token struct {
	AccessToken string `json:"access_token"`
}

type AccessToken struct {
	Raw    string `json:"rawToken"`
	Claims accessClaims
}

type RefreshToken struct {
	Raw    string `json:"rawToken"`
	Claims refreshClaims
}

type refreshClaims struct {
	SessionID string `json:"sessionID"`
	jwt.RegisteredClaims
}

type accessClaims struct {
	SessionID string `json:"sessionID"`
	jwt.RegisteredClaims
}

// New creates JWT client.
func New(accessSecretPrivate, accessSecretPublic, refreshSecret []byte,
	accessExpires, refreshExpires time.Duration) JWT {
	return JWT{
		accessSecretPrivate: accessSecretPrivate,
		accessSecretPublic:  accessSecretPublic,
		refreshSecret:       refreshSecret,
		accessExpires:       accessExpires,
		refreshExpires:      refreshExpires,
	}
}

func NewFx(cfg *config.Config) JWT {
	return New(
		[]byte(cfg.JWT.AccessSecretPrivate),
		[]byte(cfg.JWT.AccessSecretPublic),
		[]byte(cfg.JWT.RefreshSecret),
		cfg.JWT.AccessExpire,
		cfg.JWT.RefreshExpire,
	)
}

// NewPair generate new pair jwt tokens.
func (j JWT) GeneratePair(userID, sessionID uuid.UUID) (model.Token, model.Token, error) {
	const op = "pkg.jwt.NewPair"

	access, err := j.newAccess(userID, sessionID)
	if err != nil {
		return model.Token{}, model.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	refresh, err := j.newRefresh(userID, sessionID)
	if err != nil {
		return model.Token{}, model.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return access, refresh, nil
}

// ParseAccess parse raw token and return claims.
func (j JWT) ParseAccess(raw string, f interface{}) (model.ParsedToken, error) {
	var key interface{}
	switch parseFunc := f.(type) {
	case func(key []byte) (*rsa.PrivateKey, error):
		key, _ = parseFunc(j.accessSecretPrivate)
	case func(key []byte) (*rsa.PublicKey, error):
		key, _ = parseFunc(j.accessSecretPublic)
	default:
		return model.ParsedToken{}, ErrInvalidParseFunc
	}

	var cl accessClaims

	_, err := jwt.ParseWithClaims(raw, &cl, func(_ *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return model.ParsedToken{}, fmt.Errorf("%w: %w", ErrExpired, err)
		}

		return model.ParsedToken{}, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	userID, err := uuid.Parse(cl.Issuer)
	if err != nil {
		return model.ParsedToken{}, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	sessionID, err := uuid.Parse(cl.SessionID)
	if err != nil {
		return model.ParsedToken{}, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	return model.ParsedToken{UserID: userID, SessionID: sessionID}, nil
}

// ParseRefresh parse raw token and return claims.
func (j JWT) ParseRefresh(raw string) (model.ParsedToken, error) {
	var cl refreshClaims

	_, err := jwt.ParseWithClaims(raw, &cl, func(_ *jwt.Token) (interface{}, error) {
		return j.refreshSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return model.ParsedToken{}, fmt.Errorf("%w: %w", ErrExpired, err)
		}

		return model.ParsedToken{}, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	userID, err := uuid.Parse(cl.Issuer)
	if err != nil {
		return model.ParsedToken{}, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	sessionID, err := uuid.Parse(cl.SessionID)
	if err != nil {
		return model.ParsedToken{}, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	return model.ParsedToken{UserID: userID, SessionID: sessionID}, nil
}

func (j JWT) newRefresh(userID, sessionID uuid.UUID) (model.Token, error) {
	const op = "jwt.newRefresh"

	claims := refreshClaims{
		SessionID: sessionID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshExpires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	raw, err := t.SignedString(j.refreshSecret)
	if err != nil {
		return model.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return model.Token{Token: raw, ExpireAt: claims.ExpiresAt.Time}, nil
}

func (j JWT) newAccess(userID, sessionID uuid.UUID) (model.Token, error) {
	const op = "jwt.newAccess"

	key, err := jwt.ParseRSAPrivateKeyFromPEM(j.accessSecretPrivate)
	if err != nil {
		return model.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	claims := accessClaims{
		SessionID: sessionID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessExpires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)

	raw, err := t.SignedString(key)
	if err != nil {
		return model.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return model.Token{Token: raw, ExpireAt: claims.ExpiresAt.Time}, nil
}
