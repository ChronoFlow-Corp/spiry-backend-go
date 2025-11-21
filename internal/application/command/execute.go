package command

import (
	"net/url"

	"github.com/google/uuid"
)

type Execute struct {
	ChatID    *uuid.UUID
	ToolName  string
	Prompt    string
	ModelName string
	Settings  map[string]string
	Flags     []string
	Media     []*url.URL
}
