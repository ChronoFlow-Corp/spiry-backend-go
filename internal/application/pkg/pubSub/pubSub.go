package pubSub

import "sync"

type PubSub[T any] struct {
	mu    sync.RWMutex
	subs  map[chan<- T]struct{}
	input <-chan T
}

func NewPubSub[T any](input <-chan T) *PubSub[T] {
	ps := &PubSub[T]{
		subs:  make(map[chan<- T]struct{}),
		input: input,
		mu:    sync.RWMutex{},
	}

	go ps.run()

	return ps
}

func (ps *PubSub[T]) Subscribe(sub chan<- T) {
	ps.mu.Lock()
	ps.subs[sub] = struct{}{}
	ps.mu.Unlock()
}

func (ps *PubSub[T]) Unsubscribe(sub chan<- T) {
	ps.mu.Lock()
	delete(ps.subs, sub)
	ps.mu.Unlock()
	close(sub) // Graceful close
}

func (ps *PubSub[T]) run() {
	defer func() {
		ps.mu.RLock()

		for sub := range ps.subs {
			close(sub)
		}

		ps.mu.RUnlock()
	}()

	for msg := range ps.input {
		ps.mu.RLock()

		for sub := range ps.subs {
			sub <- msg
		}

		ps.mu.RUnlock()
	}
}
