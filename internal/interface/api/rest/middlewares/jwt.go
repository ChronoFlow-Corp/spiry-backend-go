package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

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
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			accessCookie, err := r.Cookie("access_token")
			if err != nil {
				slctx.Logger(r.Context()).Debug("No access token cookie", slog.Any("error", err))

				pkg.RespondError(w, http.StatusUnauthorized, response.Error{
					Code:    http.StatusUnauthorized,
					Message: "Access token required",
				})
				return
			}

			err = accessCookie.Valid()
			if err != nil {
				slctx.Logger(r.Context()).Debug("Access token is invalid", slog.Any("error", err))

				pkg.RespondError(w, http.StatusUnauthorized, response.Error{
					Code:    http.StatusUnauthorized,
					Message: "Access token is invalid",
				})
			}

			raw := accessCookie.Value
			if raw == "" {
				slctx.Logger(r.Context()).Debug("No Authorization header")
				pkg.RespondError(
					w,
					http.StatusUnauthorized,
					response.Error{Message: "Authorization required"},
				)
				return
			}

			token, err := j.ParseAccess(raw, exJwt.ParseRSAPublicKeyFromPEM)
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
					slog.String("token", raw),
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
			r = r.WithContext(addTokenCtx(r.Context(), j, w, r))

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
	accessCookie, err := r.Cookie("access_token")
	if err != nil {
		slctx.Logger(ctx).Debug("No access token cookie", slog.Any("error", err))

		return ctx
	}

	err = accessCookie.Valid()
	if err != nil {
		slctx.Logger(ctx).Debug("Invalid cookie", slog.Any("error", err))

		pkg.RespondError(w, http.StatusUnauthorized, response.Error{Message: "Invalid cookie"})
	}

	raw := accessCookie.Value
	if raw == "" {
		slctx.Logger(r.Context()).Debug("No Authorization header")
		pkg.RespondError(
			w,
			http.StatusUnauthorized,
			response.Error{Message: "Authorization required"},
		)
		return ctx
	}

	token, err := j.ParseAccess(raw, exJwt.ParseRSAPublicKeyFromPEM)
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
			slog.String("token", raw),
			slog.Any("err", err))

		return ctx
	}

	ctx = context.WithValue(r.Context(), models.UserIDCtxKey{}, token.UserID)
	ctx = context.WithValue(ctx, models.SessionIDCtxKey{}, token.SessionID)

	return ctx
}
