package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/commands"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service/chatting"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/rate"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
	"github.com/coder/websocket"
	exJwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	generating = "generating"
	done       = "done"
)

type chatProvider interface {
	SendMessage(ctx context.Context, cm commands.SendMessage) (*chatting.Socket, error)
}

type jwtProvider interface {
	ParseAccess(raw string, f interface{}) (jwt.AccessToken, error)
}

func New(chat chatProvider, j jwtProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{Subprotocols: []string{"json"}})
		if err != nil {
			tr.RespondError(w, http.StatusBadRequest, errorResponse{
				ErrorCode: "400",
				Message:   "cannot open websocket connection",
			})

			return
		}
		id := getUserID(r, j)

		if c.Subprotocol() != "json" {
			c.Close(websocket.StatusPolicyViolation, "Invalid subprotocol")

			return
		}

		defer c.Close(websocket.StatusGoingAway, "End")

		l := rate.NewLimiter(time.Millisecond*100, 10)
		log := slctx.Logger(r.Context()).With(slog.String("component", "ws_handler"))
		for {
			err := handlerMessage(c, l, log, chat, id)
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				slctx.Logger(r.Context()).Debug("websocket connection closed")
				return
			}
			if err != nil {
				slctx.Logger(r.Context()).Debug("websocket handler error", slog.String("error", err.Error()))
				return
			}
		}

	}
}

func handlerMessage(
	c *websocket.Conn,
	l *rate.Limiter,
	log *slog.Logger,
	chat chatProvider,
	userID *uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if userID != nil {
		log.Debug("handling new message", slog.Any("user_id", userID))
		ctx = context.WithValue(ctx, entities.UserIDCtxKey{}, *userID)
	}

	l.Wait(ctx)

	mstp, msg, err := c.Reader(ctx)
	if err != nil {
		log.Debug("cannot get reader", slog.Any("error", err))

		return err
	}

	var cm request

	err = json.NewDecoder(msg).Decode(&cm)
	if err != nil {
		log.Debug("cannot decode message", slog.Any("error", err))

		return err
	}

	ctx = slctx.WithLogger(ctx, log)

	res, err := chat.SendMessage(ctx, commands.SendMessage{
		ChatID: cm.ChatID,
		Role:   entities.UserRole,
		Flags:  cm.Flags,
		Vars:   cm.Vars,
		Tool:   cm.Tool,
	})
	if err != nil {
		log.Debug("cannot send message", slog.Any("error", err))

		var errNotFound *repository.ErrorNotFound
		if errors.As(err, &errNotFound) {
			raw, err := json.Marshal(errorResponse{
				ErrorCode: strconv.Itoa(http.StatusBadRequest),
				Message: fmt.Sprintf(
					"Cannot send message, not found %s with value %s",
					errNotFound.RowName,
					errNotFound.Row,
				),
			})
			if err != nil {
				log.Debug("cannot marshal message", slog.Any("error", err))
				return err
			}
			c.Write(ctx, mstp, raw)
		}

		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev := <-res.InfoEvent():
			raw, err := json.Marshal(responseChunk{
				Content:   ev.Content,
				MessageID: ev.ID.String(),
				Role:      ev.Role,
				State:     generating,
				ChatID:    ev.ChatID.String()})
			if err != nil {
				return err
			}
			c.Write(ctx, mstp, raw)
		case ev := <-res.WarningEvent():
			log.Debug("websocket warning", slog.Any("warn", ev))
			raw, err := json.Marshal(responseChunk{State: done, MessageID: res.GetMessageID().String()})
			if err != nil {
				return err
			}
			c.Write(ctx, mstp, raw)
		case err := <-res.ErrorEvent():
			log.Debug("websocket handler error", slog.Any("error", err))

			raw, err := json.Marshal(errorResponse{
				ErrorCode: "500",
				Message:   err.Error(),
			})
			if err != nil {
				return err
			}

			c.Write(ctx, mstp, raw)
		}
	}

}

func getUserID(r *http.Request, j jwtProvider) *uuid.UUID {
	raw := r.Header.Get("Authorization")
	if raw == "" {
		return nil
	}

	bearerToken := strings.TrimPrefix(raw, "Bearer ")

	token, err := j.ParseAccess(bearerToken, exJwt.ParseRSAPublicKeyFromPEM)
	if err != nil {
		return nil
	}

	id, err := uuid.Parse(token.Claims.Issuer)
	if err != nil {
		return nil
	}

	return &id
}

type request struct {
	ChatID *uuid.UUID        `json:"chat_id,omitempty"`
	Flags  []string          `json:"flags,omitempty"`
	Vars   map[string]string `json:"vars,omitempty"`
	Tool   string            `json:"tool,omitempty"`
	Media  io.ReadCloser     `json:"media,omitempty"`
}

type responseChunk struct {
	MessageID string `json:"message_id,omitempty"`
	Content   string `json:"content,omitempty"`
	ChatID    string `json:"chat_id,omitempty"`
	Role      string `json:"role,omitempty"`
	State     string `json:"state,omitempty"`
}

type errorResponse struct {
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
}
