package deleteChat

import (
	"context"
	"net/http"

	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
	"github.com/google/uuid"
)

type chatProvider interface{
	DeleteChat(ctx context.Context, chatID uuid.UUID) error
}

func New(cp chatProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawChatID := r.URL.Query().Get("chatID")

		chatID, err := uuid.Parse(rawChatID)
		if err != nil {
			tr.RespondError(w, http.StatusBadRequest, errorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid chat ID",
			} )

			return
		}

		err = cp.DeleteChat(context.Background(), chatID)
		if err != nil {
			tr.RespondError(w, http.StatusInternalServerError, errorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
			})

			return
		}
	}
}


type errorResponse struct{
	Code int `json:"code"`
	Message string `json:"message"`
}
