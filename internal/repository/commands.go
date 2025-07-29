package repository

import "github.com/google/uuid"

type ChatType = string

const (
	Default ChatType = "default"
	CaptionGen = "captionGen"
)
type ChatCommand struct {
	ChatID string `json:"chatID,omitempty"`
	UserID *uuid.UUID
	Type     ChatType `json:"type,omitempty"`
	Text     string   `json:"text,omitempty"`
	Network  string   `json:"network,omitempty"`
	Language string   `json:"language,omitempty"`
}

func (c ChatCommand) GetVars() map[string]string {
	return map[string]string{
		"social_network": c.Network,
		"language":       c.Language,
	}
}