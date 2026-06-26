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

// maxStates hard-caps the store so a flood of /auth/google/login calls can't
// grow it without bound (in addition to the per-IP rate limit on that route).
const maxStates = 10000

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
	// Hard cap: if still over budget after pruning expired entries, evict the
	// soonest-to-expire ones to make room (bounded extra work).
	for len(s.m) >= maxStates {
		var oldestKey string
		var oldest time.Time
		for k, e := range s.m {
			if oldestKey == "" || e.expiresAt.Before(oldest) {
				oldestKey, oldest = k, e.expiresAt
			}
		}
		delete(s.m, oldestKey)
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
