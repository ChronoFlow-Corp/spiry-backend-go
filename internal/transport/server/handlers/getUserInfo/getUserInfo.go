package getUserInfo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
)

type authProvider interface {
	GetUserInfo(ctx context.Context) (entities.User, error)
}

func New(a authProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := a.GetUserInfo(r.Context())
		if err != nil {
			var dbNotFound *repository.ErrorNotFound
			if errors.As(err, &dbNotFound) {
				tr.RespondError(w, http.StatusUnauthorized, errorResponse{
					Message: fmt.Sprintf("id(%s) not found", dbNotFound.RowName),
				})
			}

			slctx.Logger(r.Context()).Debug("Unknown error getting user info", slog.Any("unknown error", err))

			return
		}

		tr.RespondOK(w, okResponse{
			Email:    user.Email,
			UserName: user.Name,
			PlanName: user.Plan.Name,
			Theme:    user.Theme,
		})
	}
}

type okResponse struct {
	Email    string `json:"email"`
	UserName string `json:"user_name"`
	PlanName string `json:"plan_name"`
	Theme    string `json:"theme"`
}
type errorResponse struct {
	Message string `json:"message"`
}
