package google

import (
	"context"
	"errors"
	"fmt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
		stateMp := make(map[string]string)
		state := r.URL.Query().Get("state")
		stateKv := strings.Split(state, "=")

		if len(stateKv)%2 != 0 {
			tr.RedirectError(w, frontendURL, http.StatusInternalServerError, "Internal server error")

			return
		}

		for i := 0; i < len(stateKv); i += 2 {
			stateMp[stateKv[i]] = stateKv[i+1]
		}

		access, refresh, err := a.Login(r.Context(), stateMp, code)
		if err != nil {
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

			return
		}

		//TODO: redirect with tokens
		q := frontendURL.Query()
		q.Set("code", strconv.Itoa(http.StatusOK))
		q.Set("accessToken", access.Raw)
		q.Set("refreshToken", refresh.Raw)
		frontendURL.RawQuery = q.Encode()
		http.Redirect(w, r, frontendURL.String(), http.StatusPermanentRedirect)
	}
}

type responseOK struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
