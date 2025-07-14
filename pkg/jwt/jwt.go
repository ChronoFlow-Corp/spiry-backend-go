package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
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
	claims accessClaims
}

type RefreshToken struct {
	Raw    string `json:"rawToken"`
	claims refreshClaims
}

type refreshClaims struct {
	jwt.RegisteredClaims
}

type accessClaims struct {
	jwt.RegisteredClaims
}

// New creates JWT client.
func New(accessSecretPrivate, accessSecretPublic, refreshSecret []byte,
	accessExpires, refreshExpires time.Duration) JWT {
	return JWT{
		accessSecretPrivate: accessSecretPrivate,
		accessSecretPublic:  accessSecretPublic,
		refreshSecret:       refreshSecret,
		accessExpires: accessExpires,
		refreshExpires: refreshExpires,
	}
}

// NewPair generate new pair jwt tokens.
func (j JWT) NewPair(userID string) (AccessToken, RefreshToken, error) {
	const op = "pkg.jwt.NewPair"

	access, err := j.newAccess(userID)
	if err != nil {
		return AccessToken{}, RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	refresh, err := j.newRefresh(userID)
	if err != nil {
		return AccessToken{}, RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	return access, refresh, nil
}

// ParseAccess parse raw token and return claims.
func (j JWT) ParseAccess(raw string, f interface{}) (AccessToken, error) {
	var key interface{}
	switch parseFunc := f.(type) {
	case func(key []byte) (*rsa.PrivateKey, error):
		key, _ = parseFunc(j.accessSecretPrivate)
	case func(key []byte) (*rsa.PublicKey, error):
		key, _ = parseFunc(j.accessSecretPublic)
	default:
		return AccessToken{}, ErrInvalidParseFunc
	}

	var cl accessClaims

	_, err := jwt.ParseWithClaims(raw, &cl, func(_ *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return AccessToken{}, fmt.Errorf("%w: %w", ErrExpired, err)
		}

		return AccessToken{}, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	return AccessToken{raw, cl}, nil
}

// ParseRefresh parse raw token and return claims.
func (j JWT) ParseRefresh(raw string) (RefreshToken, error) {
	var cl refreshClaims

	_, err := jwt.ParseWithClaims(raw, &cl, func(_ *jwt.Token) (interface{}, error) {
		return j.refreshSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return RefreshToken{}, fmt.Errorf("%w: %w", ErrExpired, err)
		}

		return RefreshToken{}, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	return RefreshToken{raw, cl}, nil
}

func (j JWT) newRefresh(userID string) (RefreshToken, error) {
	const op = "jwt.newRefresh"

	claims := refreshClaims{
		jwt.RegisteredClaims{
			Issuer:    userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshExpires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	raw, err := t.SignedString(j.refreshSecret)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	return RefreshToken{raw, claims}, nil
}

func (j JWT) newAccess(userID string) (AccessToken, error) {
	const op = "jwt.newAccess"

	key, err := jwt.ParseRSAPrivateKeyFromPEM(j.accessSecretPrivate)
	if err != nil {
		return AccessToken{}, fmt.Errorf("%s: %w", op, err)
	}

	claims := accessClaims{
		jwt.RegisteredClaims{
			Issuer:    userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessExpires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)

	raw, err := t.SignedString(key)
	if err != nil {
		return AccessToken{}, fmt.Errorf("%s: %w", op, err)
	}

	return AccessToken{raw, claims}, nil
}
