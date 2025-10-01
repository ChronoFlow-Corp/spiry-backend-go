package getChats

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
	"github.com/google/uuid"
)

type chatProvider interface {
	GetChats(ctx context.Context, chatID *uuid.UUID) ([]entities.Chat, error)
}

func New(cp chatProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chatIDraw := r.URL.Query().Get("chatID")

		var ch []entities.Chat

		var err error

		var chatID uuid.UUID

		if chatIDraw == "" {
			ch, err = cp.GetChats(r.Context(), nil)
			if err != nil {
				slctx.Logger(r.Context()).Debug("GetChats err: ", slog.Any("err", err))
				tr.RespondError(w, http.StatusInternalServerError, "GetChats err: "+err.Error())

				return
			}
		} else {
			chatID, err = uuid.Parse(chatIDraw)
			if err != nil {
				slctx.Logger(r.Context()).Debug("Error parsing chatID")
				tr.RespondError(w, http.StatusBadRequest, errResponse{
					Code:    http.StatusBadRequest,
					Message: "Invalid chat ID",
				})
			}

			ch, err = cp.GetChats(r.Context(), &chatID)
			if err != nil {
				slctx.Logger(r.Context()).Debug("Error getting chat")
				tr.RespondError(w, http.StatusInternalServerError, errResponse{
					Code:    http.StatusInternalServerError,
					Message: "Error getting chat",
				})
			}
		}

		if len(ch) == 0 {
			slctx.Logger(r.Context()).Debug("No chat found")
			tr.RespondOK(w, []response{})

			return
		}

		rs := make([]response, 0, len(ch))

		for _, chat := range ch {
			tmpRs := response{}
			tmpRs.Title = chat.Title
			tmpRs.ID = chat.ID
			tmpMsgs := make([]message, 0)

			for _, msg := range chat.Messages {
				tmpMsgs = append(tmpMsgs, message{
					ID:     msg.ID,
					Content: msg.Text,
					Role:   msg.Role,
					CreatedAt: msg.CreatedAt,
					UpdatedAt: msg.UpdatedAt,
					ChatID:    msg.ChatID,
					UserID:    msg.UserID,
				})
			}

			tmpRs.Message = tmpMsgs
			rs = append(rs, tmpRs)
		}

		tr.RespondOK(w, rs)
	}
}

type response struct {
	ID      uuid.UUID `json:"id"`
	Title   string    `json:"title"`
	Message []message `json:"message,omitempty"`
}

type message struct {
	ID        uuid.UUID `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	Role string
	UpdatedAt time.Time `json:"updatedAt"`
	ChatID    uuid.UUID `json:"chatID"`
	UserID    uuid.UUID `json:"userID"`
}

type errResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
