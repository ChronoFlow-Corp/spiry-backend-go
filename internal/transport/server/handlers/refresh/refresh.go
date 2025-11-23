package refresh

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
)

type auth interface {
	Refresh(ctx context.Context, token string) (jwt.AccessToken, jwt.RefreshToken, error)
}

func New(a auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			slctx.Logger(r.Context()).Debug("Refresh error", slog.Any("error", err))
			tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Cookie not found"})

			return
		}

		access, refresh, err := a.Refresh(r.Context(), cookie.Value)
		if err != nil {
			slctx.Logger(r.Context()).Debug("Refresh error", slog.Any("error", err))
			tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Refresh failed"})

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

		tr.RespondOK(w, response{
			Access: access.Raw,
		})
	}
}

type response struct {
	Access string `json:"access"`
}

type errorResponse struct {
	Message string `json:"message"`
}
