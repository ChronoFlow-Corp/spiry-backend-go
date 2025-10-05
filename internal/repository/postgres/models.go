package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
	"github.com/google/uuid"
)

type user struct {
	ID                 uuid.UUID `db:"id"`
	Email              string    `db:"email"`
	Name               string    `db:"name"`
	Picture            string    `db:"picture"`
	AccessTokenGoogle  string    `db:"access_token_google"`
	RefreshTokenGoogle string    `db:"refresh_token_google"`
	RefreshToken       string    `db:"refresh_token"`
	Admin              bool      `db:"admin"`
	Theme              string    `db:"theme"`
	Language           string    `db:"language"`
	Plan               dbPlan    `db:"plan"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

func toUser(u user) entities.User {
	return entities.NewUser(
		u.ID,
		u.Email,
		u.Name,
		u.Picture,
		u.AccessTokenGoogle,
		u.RefreshTokenGoogle,
		u.RefreshToken,
		u.Language,
		u.Admin,
		dbPlanToRepositoryPlan(u.Plan),
		u.Theme,
	)
}

func dbChatToRepositoryChat(chat dbChat) entities.Chat {
	msgs := []entities.Message{}
	if len(chat.Messages) > 0 {
		for _, msg := range chat.Messages {
			msgs = append(msgs, entities.Message{
				ID:        msg.MessageID,
				Text:      msg.Content,
				Role:      msg.Role,
				ChatID:    msg.ChatID,
				CreatedAt: msg.CreatedAt,
				UpdatedAt: msg.UpdatedAt,
			})
		}
	} else {
		msgs = nil
	}

	return entities.Chat{
		ID:       chat.ID,
		Title:    chat.Title,
		UserID:   chat.UserID,
		Messages: msgs,
	}
}

type dbChat struct {
	ID        uuid.UUID   `db:"id"`
	Title     string      `db:"title"`
	UserID    uuid.UUID   `db:"user_id"`
	CreatedAt time.Time   `db:"created_at"`
	Messages  []dbMessage `db:"messages"`
	ToolID    uuid.UUID   `db:"tool_id"`
	UpdatedAt time.Time   `db:"updated_at"`
}

type dbChatWithMessage struct {
	ID           uuid.UUID  `db:"id"`
	Title        string     `db:"title"`
	ToolID       string     `db:"tool_id"`
	UserID       uuid.UUID  `db:"user_id"`
	CreatedAt    time.Time  `db:"chat_created_at"`
	UpdatedAt    time.Time  `db:"chat_updated_at"`
	MessageID    *uuid.UUID `db:"msg_id"`
	Content      string     `db:"content"`
	Role         string     `db:"role"`
	ChatID       uuid.UUID  `db:"chat_id"`
	MsgUserID    uuid.UUID  `db:"msg_user_id"`
	MsgUpdatedAt time.Time  `db:"msg_updated_at"`
	MsgCreatedAt time.Time  `db:"msg_created_at"`
}

type dbMessage struct {
	MessageID uuid.UUID `db:"message.id"`
	Content   string    `db:"content"`
	Role      string    `db:"role"`
	ChatID    uuid.UUID `db:"chat_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type dbTool struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Model     string    `db:"model"`
	Prompt    string    `db:"prompt"`
	Vars      dbVars    `db:"vars"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type dbVars map[string]string

func (d dbVars) Scan(val any) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, &d)
	case string:
		return json.Unmarshal([]byte(v), &d)
	default:
		return errors.New(fmt.Sprintf("Unsupported type: %T", v))
	}
}

func (d dbVars) Value() (any, error) {
	if len(d) == 0 {
		return "{}", nil
	}

	return json.Marshal(d)
}

func dbToolToRepositoryTool(tool dbTool) entities.Tool {
	return entities.Tool{
		ID:        tool.ID,
		Name:      tool.Name,
		Model:     tool.Model,
		Prompt:    tool.Prompt,
		Vars:      tool.Vars,
		CreatedAt: tool.CreatedAt,
		UpdatedAt: tool.UpdatedAt,
	}
}

type dbPlan struct {
	ID        uuid.UUID  `db:"id"`
	Name      string     `db:"name"`
	Price     string     `db:"price"`
	Limit     *int       `db:"prompt_limit"`
	End       *time.Time `db:"end_date"`
	Features  *[]string  `db:"features"`
	UserID    uuid.UUID  `db:"user_id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

func dbPlanToRepositoryPlan(plan dbPlan) entities.Plan {
	if plan.Features == nil {
		plan.Features = new([]string)
	}

	return entities.Plan{
		ID:        plan.ID,
		Name:      plan.Name,
		Price:     plan.Price,
		Limit:     plan.Limit,
		Features:  *plan.Features,
		End:       plan.End,
		UserID:    plan.UserID,
		CreatedAt: plan.CreatedAt,
		UpdatedAt: plan.UpdatedAt,
	}
}
