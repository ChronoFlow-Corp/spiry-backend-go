package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
	"github.com/go-chi/chi/v5/middleware"
	exJwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type jwtProvider interface {
	ParseAccess(raw string, f interface{}) (jwt.AccessToken, error)
}

func AuthJwt(j jwtProvider) func(next http.Handler) http.Handler {
	const op = "transport.middlewares.AuthJwt"
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("Authorization")
			if raw == "" {
				slctx.Logger(r.Context()).Debug("No Authorization header")
				tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Authorization required"})
				return
			}
			if strings.HasPrefix(raw, "Bearer ") != true {
				slctx.Logger(r.Context()).Debug("Authorization header does not start with Bearer")
				tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Authorization required"})
				return
			}

			rawToken := strings.TrimPrefix(raw, "Bearer ")
			if rawToken == "" {
				slctx.Logger(r.Context()).Debug("No token provided after Bearer")
				tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Authorization required"})
				return
			}

			token, err := j.ParseAccess(rawToken, exJwt.ParseRSAPublicKeyFromPEM)
			if err != nil {
				if errors.Is(err, jwt.ErrExpired) {
					slctx.Logger(r.Context()).Debug("Token expired")
					tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Token expired"})
				}

				if errors.Is(err, jwt.ErrInvalid) {
					slctx.Logger(r.Context()).Debug("Token invalid")
					tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Token invalid"})
				}

				slctx.Logger(r.Context()).Error("Token parse error",
					slog.String("token", rawToken),
					slog.Any("err", err))

				return
			}

			userID, err := uuid.Parse(token.Claims.Issuer)
			if err != nil {
				slctx.Logger(r.Context()).Error("Token parse error", slog.Any("err", err))
				tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "Token parse error"})
			}

			ctx := context.WithValue(r.Context(), entities.UserIDCtxKey{}, userID)
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
