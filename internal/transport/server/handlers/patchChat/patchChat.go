package patchChat

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/commands"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
	"github.com/google/uuid"
)

type chatProvider interface {
	UpdateChat(ctx context.Context, chat commands.UpdateChatTitle) error
}

func New(ch chatProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body requestBody

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			slctx.Logger(r.Context()).Debug("Body decode error", slog.Any("error", err))
			tr.RespondError(w, http.StatusBadRequest, errResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
			})
		}

		err = ch.UpdateChat(r.Context(), commands.UpdateChatTitle{
			ChatID: body.ChatID,
			NewTitle:  body.Title,
		})

		if err != nil {
			slctx.Logger(r.Context()).Debug("Chat update error", slog.Any("error", err))

			return
		}
	}
}

type requestBody struct {
	ChatID uuid.UUID `json:"chatID"`
	Title  string    `json:"title"`
}

type errResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
