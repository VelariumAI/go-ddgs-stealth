package goddgs

import (
	"fmt"
	"sync"
	"time"
)

// SessionHub tracks named fetcher sessions with TTL-based cleanup.
type SessionHub struct {
	mu       sync.Mutex
	ttl      time.Duration
	sessions map[string]*sessionEntry
}

type sessionEntry struct {
	Fetcher   Fetcher
	LastTouch time.Time
}

func NewSessionHub(ttl time.Duration) *SessionHub {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	return &SessionHub{ttl: ttl, sessions: map[string]*sessionEntry{}}
}

func (h *SessionHub) Put(name string, fetcher Fetcher) error {
	if fetcher == nil {
		return fmt.Errorf("fetcher is nil")
	}
	if name == "" {
		return fmt.Errorf("name is required")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[name] = &sessionEntry{Fetcher: fetcher, LastTouch: time.Now()}
	return nil
}

func (h *SessionHub) Get(name string) (Fetcher, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.evictExpiredLocked()
	e, ok := h.sessions[name]
	if !ok {
		return nil, false
	}
	e.LastTouch = time.Now()
	return e.Fetcher, true
}

func (h *SessionHub) Delete(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions, name)
}

func (h *SessionHub) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.evictExpiredLocked()
	return len(h.sessions)
}

func (h *SessionHub) evictExpiredLocked() {
	now := time.Now()
	for k, v := range h.sessions {
		if now.Sub(v.LastTouch) > h.ttl {
			delete(h.sessions, k)
		}
	}
}
