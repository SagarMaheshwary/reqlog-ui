package limiter

import (
	"sync/atomic"
)

type Limiter struct {
	ch     chan struct{}
	active atomic.Int64
	limit  int
}

func New(limit int) *Limiter {
	if limit <= 0 {
		limit = 10
	}
	return &Limiter{
		ch:    make(chan struct{}, limit),
		limit: limit,
	}
}

func (l *Limiter) TryAcquire() bool {
	select {
	case l.ch <- struct{}{}:
		l.active.Add(1)
		return true
	default:
		return false
	}
}

func (l *Limiter) Release() {
	select {
	case <-l.ch:
		l.active.Add(-1)
	default:
	}
}

func (l *Limiter) Active() int {
	return int(l.active.Load())
}

func (l *Limiter) Limit() int {
	return l.limit
}
