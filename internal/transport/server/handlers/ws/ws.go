package ws

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/llm"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/rate"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/tr"
	"github.com/coder/websocket"
	exJwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type chatProvider interface {
	DefaultChat(ctx context.Context, cm repository.ChatCommand) (llm.ChunkReaderCloser, *service.WsError)
}

type jwtProvider interface {
	ParseAccess(raw string, f interface{}) (jwt.AccessToken, error)
}

func New(log *slog.Logger, chat chatProvider, j jwtProvider) http.HandlerFunc {
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
		for {
			err := handlerMessage(c, l, log, chat, id)
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				log.Debug("websocket closed")
				return
			}
			if err != nil {
				log.Debug("websocket errored", slog.String("error", err.Error()))
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
	ctx, _ := context.WithTimeout(context.Background(), time.Second*40)

	l.Wait(ctx)

	mstp, msg, err := c.Reader(ctx)
	if err != nil {
		log.Debug(err.Error())

		return err
	}

	var cm request

	err = json.NewDecoder(msg).Decode(&cm)
	if err != nil {
		log.Debug("cannot decode message", slog.Attr{"error", slog.StringValue(err.Error())})

		return err
	}

	res, wsErr := chat.DefaultChat(ctx, repository.ChatCommand{
		ChatID:   cm.ChatID,
		UserID:   userID,
		Type:     cm.Type,
		Text:     cm.Text,
		Network:  cm.Network,
		Language: cm.Language,
	})

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err = wsErr.Error()
				w, wrErr := c.Writer(ctx, mstp)
				if wrErr != nil {
					log.Debug("websocket errored", slog.String("error", err.Error()))
					return
				}
				var notFound *repository.ErrorNotFound
				var uniq *repository.ErrorUnique
				var errRp errorResponse
				switch {
				case errors.As(err, &notFound):
					errRp.ErrorCode = strconv.Itoa(http.StatusNotFound)
					errRp.Message = err.Error()
					raw, _ := json.Marshal(errRp)
					w.Write(raw)
				case errors.As(err, &uniq):
					errRp.ErrorCode = strconv.Itoa(http.StatusUnprocessableEntity)
					errRp.Message = err.Error()
					raw, _ := json.Marshal(errRp)
					w.Write(raw)
				}
				w.Close()
			}
		}
	}()
	defer res.Close()

	if err != nil {
		return err
	}

	for {
		ch, err := res.Chunk()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		w, err := c.Writer(ctx, mstp)
		if err != nil {
			w.Close()
			return err
		}

		for _, v := range ch.Messages {
			raw, err := json.Marshal(responseChunk{
				Content: v.Content,
				Role:    v.Role,
				ChatID:  cm.ChatID,
			})
			if err != nil {
				log.Debug("marshaling error", slog.String("error", err.Error()))
				w.Close()

				return err
			}

			w.Write(raw)
		}

		w.Close()
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
	ChatID   string `json:"chatID,omitempty"`
	Type     string `json:"type,omitempty"`
	Text     string `json:"text,omitempty"`
	Network  string `json:"network,omitempty"`
	Language string `json:"language,omitempty"`
}

type responseChunk struct {
	Content string `json:"content"`
	ChatID  string `json:"chatID"`
	Role    string `json:"role"`
}

type errorResponse struct {
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
}
