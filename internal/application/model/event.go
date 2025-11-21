package model

import (
	"github.com/google/uuid"
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
		Type:     "generating",
		ChatID:   chatID,
		Content:  content,
		ResultID: resultID,
	}
}

func NewTitleEvent(chatID, resultID uuid.UUID, content string) Event {
	return Event{
		Type:     "title",
		ChatID:   chatID,
		Content:  content,
		ResultID: resultID,
	}
}

func NewErrorEvent(chatID, resultID uuid.UUID, cause error) Event {
	return Event{
		Type:     "error",
		ChatID:   chatID,
		Cause:    cause,
		ResultID: resultID,
	}
}

func NewWarningEvent(chatID, resultID uuid.UUID, content string) Event {
	return Event{
		Type:     "warning",
		ChatID:   chatID,
		Content:  content,
		ResultID: resultID,
	}
}

func NewDoneEvent(chatID, resultID uuid.UUID) Event {
	return Event{
		Type:     "done",
		Content:  "DONE",
		ChatID:   chatID,
		ResultID: resultID,
	}
}
