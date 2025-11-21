package rate

import (
	"context"
	"time"
)

type Limiter struct {
	t     *time.Ticker
	limit int64
	qoute chan struct{}
	done  chan struct{}
}

func NewLimiter(every time.Duration, limit int) *Limiter {
	ticker := time.NewTicker(every)
	limiter := &Limiter{
		t:     ticker,
		limit: int64(limit),
		qoute: make(chan struct{}, limit),
		done:  make(chan struct{}),
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				limiter.qoute <- struct{}{}
			case <-limiter.done:
				return
			}
		}
	}()

	return limiter
}

func (l *Limiter) Wait(ctx context.Context) {
	for {
		select {
		case <-l.qoute:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (l *Limiter) Close() {
	l.t.Stop()
}
