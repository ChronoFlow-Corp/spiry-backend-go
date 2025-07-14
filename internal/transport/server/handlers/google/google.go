package google

import (
	"context"
	"fmt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"net/http"
)

type authProvider interface {
	GetAuthCodeURI() string
	Login(ctx context.Context, state map[string]string, code string) (jwt.AccessToken, jwt.RefreshToken, error)
}

func NewRedirect(a authProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, a.GetAuthCodeURI(), http.StatusTemporaryRedirect)
	}
}

func NewCallback(a authProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		fmt.Println(state)
		st := make(map[string]string)
		a.Login(r.Context(), st, code)
	}
}
