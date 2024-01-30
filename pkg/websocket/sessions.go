package websocket

import "github.com/samber/lo"

type ISessions interface {
	Len() int
	Get(sessionID SessionID) *Session
	MGet(sessionIDs ...SessionID) ISessions
	HasKey(sessionID SessionID) bool
	Filter(fn FilterFunc) ISessions
	Range(handler func(session *Session))
	RangeWithError(handler func(session *Session) error) error
	IDs() []SessionID
}

type Sessions map[SessionID]*Session

var _ ISessions = Sessions{}

func (s Sessions) Len() int {
	return len(s)
}

func (s Sessions) Get(sessionID SessionID) *Session {
	return s[sessionID]
}

func (s Sessions) MGet(sessionIDs ...SessionID) ISessions {
	var newSessions = make(Sessions)
	for _, sessionID := range sessionIDs {
		if session, ok := s[sessionID]; ok {
			newSessions[sessionID] = session
		}
	}

	return newSessions
}

func (s Sessions) HasKey(sessionID SessionID) bool {
	_, ok := s[sessionID]
	return ok
}

func (s Sessions) Filter(fn FilterFunc) ISessions {
	if fn == nil {
		return s
	}

	filtered := make(Sessions)
	for sessionID, session := range s {
		if fn(session) {
			filtered[sessionID] = session
		}
	}

	return filtered
}

func (s Sessions) Range(handler func(session *Session)) {
	for _, session := range s {
		handler(session)
	}
}

func (s Sessions) RangeWithError(handler func(session *Session) error) error {
	for _, session := range s {
		if err := handler(session); err != nil {
			return err
		}
	}
	return nil
}

func (s Sessions) IDs() []SessionID {
	return lo.MapToSlice(s, func(sessionID SessionID, _ *Session) SessionID {
		return sessionID
	})
}
