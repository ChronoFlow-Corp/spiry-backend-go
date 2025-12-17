package pkg

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/dto/response"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/coder/websocket"
)

func RespondError(w http.ResponseWriter, code int, message response.Error) {
	raw, err := json.Marshal(message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(raw)
}

func RedirectError(w http.ResponseWriter, location *url.URL, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	q := location.Query()
	q.Set("code", strconv.Itoa(status))
	q.Set("error", message)
	location.RawQuery = q.Encode()

	w.Header().Set("Location", location.String())
	w.WriteHeader(http.StatusPermanentRedirect)
}

func RespondErrorWs(ctx context.Context, conn *websocket.Conn, mstp websocket.MessageType, rs response.ResponseChunk) {
	raw, err := json.Marshal(rs)
	if err != nil {
		slctx.Logger(ctx).Error("cannot marshal reponseChunk", slog.Any("error", err))

		return
	}

	err = conn.Write(ctx, mstp, raw)
	if err != nil {
		slctx.Logger(ctx).Error("cannot write to socket", slog.Any("error", err))

		return
	}
}
