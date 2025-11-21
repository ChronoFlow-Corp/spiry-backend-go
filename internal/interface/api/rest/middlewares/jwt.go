package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/model"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/auth/jwt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/dto/response"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/pkg"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/go-chi/chi/v5/middleware"
	exJwt "github.com/golang-jwt/jwt/v5"
)

type jwtProvider interface {
	ParseAccess(raw string, f interface{}) (model.ParsedToken, error)
}

func AuthJwt(j jwtProvider) func(next http.Handler) http.Handler {
	const op = "transport.middlewares.AuthJwt"
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("Authorization")
			if raw == "" {
				slctx.Logger(r.Context()).Debug("No Authorization header")
				pkg.RespondError(
					w,
					http.StatusUnauthorized,
					response.Error{Message: "Authorization required"},
				)
				return
			}
			if !strings.HasPrefix(raw, "Bearer ") {
				slctx.Logger(r.Context()).Debug("Authorization header does not start with Bearer")
				pkg.RespondError(
					w,
					http.StatusUnauthorized,
					response.Error{Message: "Authorization required"},
				)
				return
			}

			rawToken := strings.TrimPrefix(raw, "Bearer ")
			if rawToken == "" {
				slctx.Logger(r.Context()).Debug("No token provided after Bearer")
				pkg.RespondError(
					w,
					http.StatusUnauthorized,
					response.Error{Message: "Authorization required"},
				)
				return
			}

			token, err := j.ParseAccess(rawToken, exJwt.ParseRSAPublicKeyFromPEM)
			if err != nil {
				if errors.Is(err, jwt.ErrExpired) {
					slctx.Logger(r.Context()).Debug("Token expired")
					pkg.RespondError(
						w,
						http.StatusUnauthorized,
						response.Error{Message: "Token expired"},
					)
				}

				if errors.Is(err, jwt.ErrInvalid) {
					slctx.Logger(r.Context()).Debug("Token invalid")
					pkg.RespondError(
						w,
						http.StatusUnauthorized,
						response.Error{Message: "Token invalid"},
					)
				}

				slctx.Logger(r.Context()).Error("Token parse error",
					slog.String("token", rawToken),
					slog.Any("err", err))

				return
			}

			ctx := context.WithValue(r.Context(), models.UserIDCtxKey{}, token.UserID)
			ctx = context.WithValue(ctx, models.SessionIDCtxKey{}, token.SessionID)
			r = r.WithContext(ctx)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

func TryAuthJwt(j jwtProvider) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("Authorization")
			if raw != "" {
				r = r.WithContext(addTokenCtx(r.Context(), j, w, r))
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

func addTokenCtx(
	ctx context.Context,
	j jwtProvider,
	w http.ResponseWriter,
	r *http.Request,
) context.Context {
	raw := r.Header.Get("Authorization")
	if raw == "" {
		slctx.Logger(r.Context()).Debug("No Authorization header")
		pkg.RespondError(
			w,
			http.StatusUnauthorized,
			response.Error{Message: "Authorization required"},
		)
		return ctx
	}
	if !strings.HasPrefix(raw, "Bearer ") {
		slctx.Logger(r.Context()).Debug("Authorization header does not start with Bearer")
		pkg.RespondError(
			w,
			http.StatusUnauthorized,
			response.Error{Message: "Authorization required"},
		)
		return ctx
	}

	rawToken := strings.TrimPrefix(raw, "Bearer ")
	if rawToken == "" {
		slctx.Logger(r.Context()).Debug("No token provided after Bearer")
		pkg.RespondError(
			w,
			http.StatusUnauthorized,
			response.Error{Message: "Authorization required"},
		)
		return ctx
	}

	token, err := j.ParseAccess(rawToken, exJwt.ParseRSAPublicKeyFromPEM)
	if err != nil {
		if errors.Is(err, jwt.ErrExpired) {
			slctx.Logger(r.Context()).Debug("Token expired")
			pkg.RespondError(
				w,
				http.StatusUnauthorized,
				response.Error{Message: "Token expired"},
			)
		}

		if errors.Is(err, jwt.ErrInvalid) {
			slctx.Logger(r.Context()).Debug("Token invalid")
			pkg.RespondError(
				w,
				http.StatusUnauthorized,
				response.Error{Message: "Token invalid"},
			)
		}

		slctx.Logger(r.Context()).Error("Token parse error",
			slog.String("token", rawToken),
			slog.Any("err", err))

		return ctx
	}

	ctx = context.WithValue(r.Context(), models.UserIDCtxKey{}, token.UserID)
	ctx = context.WithValue(ctx, models.SessionIDCtxKey{}, token.SessionID)

	return ctx
}
