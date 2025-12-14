package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/command"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/model"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/query"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/service/chatting"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/dto/request"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/dto/response"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/middlewares"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/pkg"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/pkg/rate"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type JWTProvider interface {
	ParseAccess(raw string, f interface{}) (model.ParsedToken, error)
}

type ChattingModule struct {
	service     chatting.Service
	JWTProvider JWTProvider
	frontendUrl *url.URL
}

func NewChattingModule(
	cfg *config.Config,
	service chatting.Service,
	j JWTProvider,
) *ChattingModule {
	u, err := url.Parse(cfg.HTTP.FrontendUrl)
	if err != nil {
		panic("invalid frontend URL " + err.Error())
	}

	return &ChattingModule{service: service, JWTProvider: j, frontendUrl: u}
}

func (m *ChattingModule) Register(r chi.Router) {
	r.Route("/chatting", func(r chi.Router) {
		r.Use(middlewares.TryAuthJwt(m.JWTProvider))
		r.Get("/ws", m.Execute)

		r.Get("/chat", m.GetChats)
		r.Patch("/chat", m.UpdateChat)
		r.Delete("/chat", m.DeleteChat)
	})
}

// GetChats godoc
//
//	@Summary	get chats list or chat history selected chat.
//	@Tags		chatting
//	@Produce	json
//	@Param		chat_id	query		string	false	"If provided return chat object with history"
//	@Success	200		{object}	[]response.Chat
//	@Failure	400		{object}	response.Error
//	@Failure	401		{object}	response.Error
//	@Failure	404		{object}	response.Error
//	@Router		/api/chatting/chat [get]
func (m *ChattingModule) GetChats(w http.ResponseWriter, r *http.Request) {
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

	res, err := m.service.GetChats(r.Context(), query.GetChats{
		ChatID: chatID,
	})
	if err != nil {
		var errNotFound *domain.ErrorNotFound
		if errors.As(err, &errNotFound) {
			pkg.RespondError(w, http.StatusNotFound, response.Error{
				Code: http.StatusNotFound,
				Message: fmt.Sprintf(
					"Chat not found with %s: %s",
					errNotFound.FieldName,
					errNotFound.FieldValue,
				),
			})

			return
		}
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
			Messages:  make([]response.Message, 0, len(chat.Couple)*2),
		}

		for _, c := range chat.Couple {
			ch.Messages = append(ch.Messages, response.Message{
				ID:        c.Command.ID,
				Text:      c.Command.Text,
				Settings:  c.Command.Settings,
				Flags:     c.Command.Flags,
				Status:    &c.Command.Status,
				Role:      response.UserRole,
				CreatedAt: c.Command.CreatedAt,
			})

			ch.Messages = append(ch.Messages, response.Message{
				ID:        c.Result.ID,
				Text:      c.Result.Text,
				CreatedAt: c.Result.CreatedAt,
				Role:      response.AssistantRole,
			})
		}

		chats = append(chats, ch)
	}

	pkg.RespondOK(w, chats)
}

// UpdateChat godoc
//
//	@Summary	update chat title
//	@Tags		chatting
//	@Accept		json
//	@Produce	json
//	@Param		chat_info	body	request.UpdateChat	true	"Payload"
//	@Success	200
//	@Failure	400	{object}	response.Error
//	@Failure	404	{object}	response.Error
//	@Router		/api/chatting/chat [patch]
func (m *ChattingModule) UpdateChat(w http.ResponseWriter, r *http.Request) {
	var req request.UpdateChat
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slctx.Logger(r.Context()).Debug("Failed to decode body", slog.Any("error", err))
		pkg.RespondError(w, http.StatusBadRequest, response.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid request body: %s", err.Error()),
		})

		return
	}

	err = m.service.UpdateTitle(
		r.Context(),
		command.UpdateTitle{Title: req.NewTitle, ChatID: req.ChatID},
	)
	if err != nil {
		slctx.Logger(r.Context()).Debug("Failed to update title", slog.Any("error", err))

		var notFound *domain.ErrorNotFound
		if errors.As(err, &notFound) {
			pkg.RespondError(w, http.StatusNotFound, response.Error{
				Code:    http.StatusNotFound,
				Message: fmt.Sprintf("%s not found: %s", notFound.FieldName, notFound.FieldValue),
			})

			return
		}
		pkg.RespondError(w, http.StatusInternalServerError, response.Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return
	}

	pkg.RespondOK(w, nil)
}

// DeleteChat godoc
//
//	@Summary	Delete chat
//	@Tags		chatting
//	@Param		chat_id	query	string	true	"ID for delete"
//
//	@Success	200
//	@Failure	400	{object}	response.Error
//	@Failure	404	{object}	response.Error
//	@Router		/api/chatting/chat [delete]
func (m *ChattingModule) DeleteChat(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("chat_id")
	if q == "" {
		pkg.RespondError(w, http.StatusBadRequest, response.Error{
			Code:    http.StatusBadRequest,
			Message: "ChatID query not provided",
		})

		return
	}

	chatID, err := uuid.Parse(q)
	if err != nil {
		pkg.RespondError(w, http.StatusBadRequest, response.Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("Invalid chat_id provided: %s", q),
		})

		return
	}

	err = m.service.DeleteChat(r.Context(), command.DeleteChat{ChatID: chatID})
	if err != nil {
		slctx.Logger(r.Context()).Debug("Failed to delete chat", slog.Any("error", err))
		var notFound *domain.ErrorNotFound
		if errors.As(err, &notFound) {
			pkg.RespondError(w, http.StatusNotFound, response.Error{
				Code:    http.StatusNotFound,
				Message: fmt.Sprintf("%s not found: %s", notFound.FieldName, notFound.FieldValue),
			})

			return
		}

		slctx.Logger(r.Context()).Error("Failed to delete chat", slog.Any("error", err))
		pkg.RespondError(w, http.StatusInternalServerError, response.Error{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
	}

	pkg.RespondOK(w, nil)
}

// Execute godoc
//
//	@Summary	Connection to ws
//	@Tags		chatting
//	@Accept		json
//	@Produce	json
//	@Param		Upgrade					header	string	true	"websocket"
//	@Param		Connection				header	string	true	"upgrade"
//	@Param		Sec-WebSocket-Protocol	header	string	true	"need access token if exist also provide bearer prefix"	example("Bearer <token>")
//	@Router		/api/chatting/ws [get]
func (m *ChattingModule) Execute(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{"json"},
		OriginPatterns: []string{
			fmt.Sprintf("%s", m.frontendUrl.Host),
		},
	})
	if err != nil {
		slctx.Logger(r.Context()).Error("websocket accept error", slog.Any("error", err))
	}

	l := rate.NewLimiter(time.Millisecond*100, 10)

	logger := slctx.Logger(r.Context())
	ctx := addCredosToCtx(r.Context())
	slctx.WithLogger(ctx, logger)
	for {
		err = m.accept(ctx, conn, l)
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			slctx.Logger(r.Context()).Debug("websocket connection closed")
			return
		}

		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusGoingAway {
				slctx.Logger(r.Context()).Debug("websocket connection closed")
				return
			}
			slctx.Logger(r.Context()).Error("websocket accept error", slog.Any("error", err))

			return
		}
	}
}

func (m *ChattingModule) accept(
	ctx context.Context,
	conn *websocket.Conn,
	l *rate.Limiter,
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
			// TODO: respond error
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
		var raw []byte
		var errNotFound *domain.ErrorNotFound
		if errors.Is(err, domain.ZeroAllowedTools) {
			raw, err = json.Marshal(response.ResponseChunk{
				Content: "There are no allowed tools",
				State:   "ERROR",
			})
			if err != nil {
				return err
			}
		}

		if errors.As(err, &errNotFound) {
			raw, err = json.Marshal(response.ResponseChunk{
				Content: fmt.Sprintf(
					"%s not found with %s",
					errNotFound.FieldName,
					errNotFound.FieldValue,
				),
				State: "ERROR",
			})
		}

		_ = conn.Write(ctx, mstp, raw)

		return err
	}

	for ch := range event.Stream() {
		switch ch.Type {
		case model.ErrorEvent:
			var raw []byte
			if errors.Is(ch.Cause, domain.ZeroAllowedTools) {
				raw, err = json.Marshal(response.ResponseChunk{
					MessageID: ch.ResultID.String(),
					Content:   "There are no allowed tools",
					ChatID:    ch.ChatID.String(),
					State:     ch.Type,
				})
				if err != nil {
					return err
				}
			}

			_ = conn.Write(ctx, mstp, raw)
		default:
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
	}

	return nil
}

func addCredosToCtx(ctx context.Context) context.Context {
	newContext := context.Background()

	userID, ok := ctx.Value(models.UserIDCtxKey{}).(uuid.UUID)
	if ok {
		newContext = context.WithValue(newContext, models.UserIDCtxKey{}, userID)
	}

	sessionID, ok := ctx.Value(models.SessionIDCtxKey{}).(uuid.UUID)
	if ok {
		newContext = context.WithValue(newContext, models.SessionIDCtxKey{}, sessionID)
	}

	return newContext
}
