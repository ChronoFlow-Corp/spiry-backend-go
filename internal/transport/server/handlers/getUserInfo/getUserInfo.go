package getUserInfo

import (
	"context"
	"errors"
	"fmt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server/middlewares"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
	"log/slog"
	"net/http"
)

type authProvider interface {
	GetUserInfo(ctx context.Context, id string) (repository.User, error)
}

func New(log *slog.Logger, a authProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := r.Context().Value(middlewares.AccessTokenKey)
		if raw == nil {
			log.Debug("No access token")
			tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "no token provided"})

			return
		}

		token, ok := raw.(string)
		if !ok {
			log.Debug("No token provided")
			tr.RespondError(w, http.StatusUnauthorized, errorResponse{Message: "no token provided"})
		}
		log.Debug("Token parsed successfully", slog.Attr{Key: "token", Value: slog.StringValue(token)})

		user, err := a.GetUserInfo(r.Context(), token)
		if err != nil {
			var dbNotFound *repository.ErrorNotFound
			if errors.As(err, &dbNotFound) {
				tr.RespondError(w, http.StatusUnauthorized, errorResponse{
					Message: fmt.Sprintf("id(%s) not found", dbNotFound.RowName),
				})
			}
			log.Debug("Error getting user info", slog.Attr{
				Key:   "unknown error",
				Value: slog.StringValue(err.Error()),
			})

			return
		}

		tr.RespondOK(w, okResponse{Email: user.Email()})
	}
}

type okResponse struct {
	Email string `json:"email"`
}
type errorResponse struct {
	Message string `json:"message"`
}
