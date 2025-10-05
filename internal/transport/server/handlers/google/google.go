package google

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
)

type authProvider interface {
	GetAuthCodeURI() string
	Login(ctx context.Context, state map[string]string, code string) (jwt.AccessToken, jwt.RefreshToken, error)
}

// NewRedirect redirect to google auth with scope permissions.
func NewRedirect(a authProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, a.GetAuthCodeURI(), http.StatusTemporaryRedirect)
	}
}

// NewCallback handle google auth answer, create user and return jwt tokens pair.
func NewCallback(frontendURL *url.URL, backendDomain string, a authProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		st, err := parseState(state)
		if err != nil {
			st = map[string]string{}
		}

		_, refresh, err := a.Login(r.Context(), st, code)
		if err != nil {
			slctx.Logger(r.Context()).Debug("Login error", slog.Any("error", err))

			var authErr *service.AuthError

			var uniqErr *repository.ErrorUnique
			switch {

			case errors.Is(err, jwt.ErrExpired):
				tr.RedirectError(w, frontendURL, http.StatusUnauthorized, "Token expired")
			case errors.As(err, &uniqErr):
				tr.RedirectError(w, frontendURL, http.StatusConflict,
					fmt.Sprintf("%s: %s field not unique", uniqErr.RowName, uniqErr.Row))
			case errors.As(err, &authErr):
				tr.RedirectError(w, frontendURL, http.StatusInternalServerError, "Internal server error")
			}

			tr.RedirectError(w, frontendURL, http.StatusInternalServerError, "Internal server error")

			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refresh.Raw,
			Path:     "/api/refresh",
			Expires:  refresh.Claims.ExpiresAt.Time,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, frontendURL.String(), http.StatusPermanentRedirect)
	}
}

func parseState(state string) (map[string]string, error) {
	stateMp := make(map[string]string)
	stateKv := strings.Split(state, "=")

	if len(stateKv)%2 != 0 {
		return nil, fmt.Errorf("invalid state format")
	}

	for i := 0; i < len(stateKv); i += 2 {
		stateMp[stateKv[i]] = stateKv[i+1]
	}

	return stateMp, nil
}
