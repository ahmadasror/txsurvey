package service

import (
	"sync"
	"time"
)

// stateStore holds short-lived OAuth state -> PKCE verifier mappings. An
// in-memory TTL map is sufficient at this scale (5-10 creators, single
// instance); a server restart only invalidates in-flight logins, which simply
// retry. Keyed by the random state value returned on the OAuth callback.
type stateStore struct {
	mu  sync.Mutex
	m   map[string]stateEntry
	ttl time.Duration
}

type stateEntry struct {
	verifier  string
	expiresAt time.Time
}

func newStateStore(ttl time.Duration) *stateStore {
	return &stateStore{m: make(map[string]stateEntry), ttl: ttl}
}

// put records a state->verifier pair and opportunistically evicts expired ones.
func (s *stateStore) put(state, verifier string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, e := range s.m {
		if now.After(e.expiresAt) {
			delete(s.m, k)
		}
	}
	s.m[state] = stateEntry{verifier: verifier, expiresAt: now.Add(s.ttl)}
}

// take consumes a state, returning its verifier and whether it was valid and
// unexpired. A state is single-use (deleted on take) to prevent replay.
func (s *stateStore) take(state string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.m[state]
	if !ok {
		return "", false
	}
	delete(s.m, state)
	if time.Now().After(e.expiresAt) {
		return "", false
	}
	return e.verifier, true
}
