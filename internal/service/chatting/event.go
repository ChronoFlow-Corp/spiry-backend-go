package chatting

import (
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/llm"
	"github.com/google/uuid"
)

const DoneCause = "DONE"

type Socket struct{
	infoEventCh chan InfoEvent
	warningEventCh chan WarningEvent
	errorEventCh chan error
	chatID uuid.UUID
	messageID uuid.UUID
	done chan struct{}
	chunker llm.ChunkReader
}

func NewSocket(chatID uuid.UUID, messageID uuid.UUID, chunker llm.ChunkReader) *Socket {
	s := &Socket{
		infoEventCh: make(chan InfoEvent),
		warningEventCh: make(chan WarningEvent),
		errorEventCh: make(chan error),
		done: make(chan struct{}),
		chatID: chatID,
		messageID: messageID,
		chunker: chunker,
	}

	return s
}

type InfoEvent struct{
	ID uuid.UUID
	ChatID uuid.UUID
	Content string
	Role string
}

type WarningEvent struct{
	Err error
	Cause string
	Description string
}


func (s *Socket) InfoEvent() <- chan InfoEvent {
	return s.infoEventCh
}

func (s *Socket) WarningEvent() <- chan WarningEvent {
	return s.warningEventCh
}

func (s *Socket) ErrorEvent() <- chan error {
	return s.errorEventCh
}

func (s *Socket) GetMessageID() uuid.UUID {
	return s.messageID
}

func (s *Socket) sendError(err error) {
	select {
	case s.errorEventCh <- err:
	case <-s.done:
	}
}

func (s *Socket) sendInfo(event InfoEvent) {
	select {
	case s.infoEventCh <- event:
	case <-s.done:
	}
}

func (s *Socket) sendWarning(event WarningEvent) {
	select {
	case s.warningEventCh <- event:
	case <-s.done:
	}
}

func (s *Socket) Close() {
	close(s.done)
	close(s.infoEventCh)
	close(s.warningEventCh)
	close(s.errorEventCh)
}
