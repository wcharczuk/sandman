package web

import (
	"context"
	"sync"
)

var (
	_ FetchSessionHandler   = (*LocalSessionCache)(nil)
	_ PersistSessionHandler = (*LocalSessionCache)(nil)
	_ RemoveSessionHandler  = (*LocalSessionCache)(nil)
)

// LocalSessionCache is a memory cache of sessions.
// It is meant to be used in tests.
type LocalSessionCache struct {
	sync.RWMutex
	Sessions map[string]*Session
}

// FetchSession is a shim to interface with the auth manager.
func (lsc *LocalSessionCache) FetchSession(_ context.Context, sessionID string) (*Session, error) {
	return lsc.Get(sessionID), nil
}

// PersistSession is a shim to interface with the auth manager.
func (lsc *LocalSessionCache) PersistSession(_ context.Context, session *Session) error {
	lsc.Upsert(session)
	return nil
}

// RemoveSession is a shim to interface with the auth manager.
func (lsc *LocalSessionCache) RemoveSession(_ context.Context, sessionID string) error {
	lsc.Remove(sessionID)
	return nil
}

// Upsert adds or updates a session to the cache.
func (lsc *LocalSessionCache) Upsert(session *Session) {
	lsc.Lock()
	defer lsc.Unlock()

	if lsc.Sessions == nil {
		lsc.Sessions = make(map[string]*Session)
	}
	lsc.Sessions[session.SessionID] = session
}

// Remove removes a session from the cache.
func (lsc *LocalSessionCache) Remove(sessionID string) {
	lsc.Lock()
	defer lsc.Unlock()
	delete(lsc.Sessions, sessionID)
}

// Get gets a session.
func (lsc *LocalSessionCache) Get(sessionID string) *Session {
	lsc.RLock()
	defer lsc.RUnlock()
	if lsc.Sessions == nil {
		return nil
	}
	if session, hasSession := lsc.Sessions[sessionID]; hasSession {
		return session
	}
	return nil
}
