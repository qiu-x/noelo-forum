package session

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const SessionCookie = "session_token"

type Sessions struct {
	sessions map[string]session
	mu       sync.Mutex
}

type session struct {
	username string
}

func NewSessions() *Sessions {
	return &Sessions{
		sessions: make(map[string]session),
		mu:       sync.Mutex{},
	}
}

func generateSessionToken() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var result string
	for range 50 {
		digit := r.Intn(10)
		result += strconv.Itoa(digit)
	}
	return result
}

func (s *Sessions) CheckAuth(sessionToken string) (string, bool) {
	// Reads from map are thread-safe
	if v, ok := s.sessions[sessionToken]; ok {
		return v.username, true
	}
	return "", false
}

// CreateSession creates a session token for an authenticated username
func (s *Sessions) CreateSession(username string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionToken := generateSessionToken()
	s.sessions[sessionToken] = session{
		username: username,
	}
	return sessionToken
}

func (s *Sessions) Logout(sessionToken string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionToken)
}
