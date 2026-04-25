package tokenstore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Store struct {
	mu     sync.Mutex
	tokens map[string]time.Time
	ttl    time.Duration
}

func New(ctx context.Context, ttl time.Duration) *Store {
	s := &Store{
		tokens: make(map[string]time.Time),
		ttl:    ttl,
	}
	go s.reap(ctx)
	return s
}

func (s *Store) Issue() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	s.mu.Lock()
	s.tokens[token] = time.Now().Add(s.ttl)
	s.mu.Unlock()

	return token, nil
}

func (s *Store) Consume(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	exp, ok := s.tokens[token]
	if !ok {
		return false
	}
	delete(s.tokens, token)
	return time.Now().Before(exp)
}

func (s *Store) reap(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			s.mu.Lock()
			for t, exp := range s.tokens {
				if now.After(exp) {
					delete(s.tokens, t)
				}
			}
			s.mu.Unlock()
		}
	}
}
