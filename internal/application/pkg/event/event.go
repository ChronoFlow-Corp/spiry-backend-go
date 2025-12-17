package event

import (
	"context"
	"errors"
	"sync"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/model"
)

var ErrClosed = errors.New("event manager closed")

type Manager struct {
	mu     sync.RWMutex
	ch     chan model.Event
	ctx    context.Context
	cancel context.CancelFunc
	closed bool
}

func NewManager(parent context.Context, buffer int) *Manager {
	ctx, cancel := context.WithCancel(parent)

	return &Manager{
		ch:     make(chan model.Event, buffer),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (m *Manager) Send(e model.Event) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return ErrClosed
	}

	select {
	case <-m.ctx.Done():
		return ErrClosed
	case m.ch <- e:
		return nil
	}
}

func (m *Manager) Closed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.closed
}

func (m *Manager) Stream() <-chan model.Event {
	return m.ch
}

func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}

	m.closed = true
	m.cancel()
	close(m.ch)
}
