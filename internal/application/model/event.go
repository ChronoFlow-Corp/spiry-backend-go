package model

import (
	"github.com/google/uuid"
)

const (
	GeneratingEvent = "generating"
	TitleEvent      = "title"
	ErrorEvent      = "error"
	WarningEvent    = "warning"
	DoneEvent       = "done"
)

type Event struct {
	Type     string
	ChatID   uuid.UUID
	Content  string
	Cause    error
	ResultID uuid.UUID
}

func NewGeneratingEvent(chatID, resultID uuid.UUID, content string) Event {
	return Event{
		Type:     GeneratingEvent,
		ChatID:   chatID,
		Content:  content,
		ResultID: resultID,
	}
}

func NewTitleEvent(chatID, resultID uuid.UUID, content string) Event {
	return Event{
		Type:     TitleEvent,
		ChatID:   chatID,
		Content:  content,
		ResultID: resultID,
	}
}

func NewErrorEvent(chatID, resultID uuid.UUID, cause error) Event {
	return Event{
		Type:     ErrorEvent,
		ChatID:   chatID,
		Cause:    cause,
		ResultID: resultID,
	}
}

func NewWarningEvent(chatID, resultID uuid.UUID, content string) Event {
	return Event{
		Type:     WarningEvent,
		ChatID:   chatID,
		Content:  content,
		ResultID: resultID,
	}
}

func NewDoneEvent(chatID, resultID uuid.UUID) Event {
	return Event{
		Type:     DoneEvent,
		Content:  "DONE",
		ChatID:   chatID,
		ResultID: resultID,
	}
}
