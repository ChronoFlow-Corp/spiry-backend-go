package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/command"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/model"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/query"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/service/chatting"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/auth/jwt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/dto/request"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/dto/response"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/pkg"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/pkg/rate"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	exJwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTProvider interface {
	ParseAccess(raw string, f interface{}) (model.ParsedToken, error)
}

type ChattingModule struct {
	service     chatting.Service
	JWTProvider JWTProvider
}

func NewChattingModule(service chatting.Service, j JWTProvider) *ChattingModule {
	return &ChattingModule{service: service, JWTProvider: j}
}

func (m *ChattingModule) Register(r chi.Router) {
	r.Get("/ws", m.Execute)

	r.Route("/chats", func(r chi.Router) {
		r.Get("/", m.GetChats)
	})
}

// GetChats godoc
func (m *ChattingModule) GetChats(w http.ResponseWriter, r *http.Request) {
	ctx := m.addTokenCtx(r.Context(), w, r)

	var chatID *uuid.UUID
	q := r.URL.Query().Get("chat_id")
	if q != "" {
		v, err := uuid.Parse(q)
		if err != nil {
			pkg.RespondError(w, http.StatusBadRequest, response.Error{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Invalid chat_id provided: %s", q),
			})
			return
		}

		chatID = &v
	}

	res, err := m.service.GetChats(ctx, query.GetChats{
		ChatID: chatID,
	})
	if err != nil {
		pkg.RespondError(w, http.StatusInternalServerError, response.Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	chats := make([]response.Chat, 0, len(res))
	for _, chat := range res {
		ch := response.Chat{
			ID:        chat.ID,
			Title:     chat.Title,
			CreatedAt: chat.CreatedAt,
			UpdatedAt: chat.UpdatedAt,
			History:   make([]response.Message, 0, len(chat.Couple)*2),
		}

		for _, c := range chat.Couple {
			ch.History = append(ch.History, response.Message{
				ID:        c.Command.ID,
				Text:      c.Command.Text,
				Settings:  c.Command.Settings,
				Flags:     c.Command.Flags,
				Status:    &c.Command.Status,
				CreatedAt: c.Command.CreatedAt,
			})

			ch.History = append(ch.History, response.Message{
				ID:        c.Result.ID,
				Text:      c.Result.Text,
				CreatedAt: c.Result.CreatedAt,
			})
		}

		chats = append(chats, ch)
	}

	pkg.RespondOK(w, chats)
}

func (m *ChattingModule) Execute(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{Subprotocols: []string{"json"}})
	if err != nil {
		slctx.Logger(r.Context()).Error("websocket accept error", slog.Any("error", err))
	}

	l := rate.NewLimiter(time.Millisecond*100, 10)

	for {
		ctx := m.addTokenCtx(context.Background(), w, r)
		//TODO: add logger

		err = m.accept(ctx, conn, l, slctx.Logger(r.Context()))
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			slctx.Logger(r.Context()).Debug("websocket connection closed")
			return
		}

		if err != nil {
			slctx.Logger(r.Context()).Error("websocket accept error", slog.Any("error", err))
			return
		}
	}
}

func (m *ChattingModule) accept(
	ctx context.Context,
	conn *websocket.Conn,
	l *rate.Limiter,
	log *slog.Logger,
) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()
	l.Wait(ctx)

	mstp, msg, err := conn.Reader(ctx)
	if err != nil {
		return err
	}

	var req request.WebSocketRequest

	err = json.NewDecoder(msg).Decode(&req)
	if err != nil {
		return err
	}

	medias := make([]*url.URL, len(req.Media))
	for i, m := range req.Media {
		u, err := url.Parse(m)
		if err != nil {
			//TODO: respond error
			return err
		}

		medias[i] = u
	}

	event, err := m.service.ExecuteCommand(ctx, command.Execute{
		ChatID:    req.ChatID,
		ToolName:  req.Tool,
		Prompt:    req.Prompt,
		ModelName: req.Model,
		Settings:  req.Vars,
		Flags:     req.Flags,
		Media:     medias,
	})
	if err != nil {
		// TODO: handle error
		return err
	}

	for ch := range event.Stream() {
		raw, err := json.Marshal(response.ResponseChunk{
			MessageID: ch.ResultID.String(),
			Content:   ch.Content,
			ChatID:    ch.ChatID.String(),
			State:     ch.Type,
		})
		if err != nil {
			return err
		}

		_ = conn.Write(ctx, mstp, raw)
	}

	return nil
}

func (m *ChattingModule) addTokenCtx(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
) context.Context {
	raw := r.Header.Get("Authorization")
	if raw == "" {
		return ctx
	}
	if !strings.HasPrefix(raw, "Bearer ") {
		return ctx
	}

	rawToken := strings.TrimPrefix(raw, "Bearer ")
	if rawToken == "" {
		return ctx
	}

	token, err := m.JWTProvider.ParseAccess(rawToken, exJwt.ParseRSAPublicKeyFromPEM)
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
