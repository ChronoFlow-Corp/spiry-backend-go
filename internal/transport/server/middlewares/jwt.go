package middlewares

import (
	"context"
	"errors"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
	"github.com/go-chi/chi/v5/middleware"
	exJwt "github.com/golang-jwt/jwt/v5"
	"log/slog"
	"net/http"
	"strings"
)

type jwtProvider interface {
	ParseAccess(raw string, f interface{}) (jwt.AccessToken, error)
}

type CtxKey string

const AccessTokenKey CtxKey = "token"

func AuthJwt(log *slog.Logger, j jwtProvider) func(next http.Handler) http.Handler {
	const op = "transport.middlewares.AuthJwt"
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("Authorization")
			if raw == "" {
				log.Debug("No Authorization header")
				tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Authorization required"})
				return
			}
			if strings.HasPrefix(raw, "Bearer ") != true {
				log.Debug("No Bearer prefix")
				tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Authorization required"})
				return
			}

			rawToken := strings.TrimPrefix(raw, "Bearer ")
			if rawToken == "" {
				log.Debug("Token required")
				tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Authorization required"})
				return
			}

			token, err := j.ParseAccess(rawToken, exJwt.ParseRSAPublicKeyFromPEM)
			if err != nil {
				if errors.Is(err, jwt.ErrExpired) {
					log.Debug("Token expired")
					tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Token expired"})
				}

				if errors.Is(err, jwt.ErrInvalid) {
					log.Debug("Token invalid")
					tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Token invalid"})
				}
				log.Debug("Token parse error", slog.Attr{Key: "Token", Value: slog.AnyValue(token)})

				return
			}
			log.Debug("Token parsed successfully", token)
			ctx := context.WithValue(r.Context(), AccessTokenKey, token)
			r = r.WithContext(ctx)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

type errorResponse struct {
	Message string `json:"message"`
}
