package web

import (
	"time"
)

// Session is an active session
type Session struct {
	UserID     string    `json:"userID,omitempty"`
	BaseURL    string    `json:"baseURL,omitempty"`
	SessionID  string    `json:"sessionID,omitempty"`
	CreatedUTC time.Time `json:"createdUTC,omitempty"`
	ExpiresUTC time.Time `json:"expiresUTC,omitempty"`
	UserAgent  string    `json:"userAgent,omitempty"`
	RemoteAddr string    `json:"remoteAddr,omitempty"`
	Locale     string    `json:"locale,omitempty"`
	State      any       `json:"state,omitempty"`
}

// IsExpired returns if the session is expired.
func (s *Session) IsExpired(asOf time.Time) bool {
	if s.ExpiresUTC.IsZero() {
		return false
	}
	return s.ExpiresUTC.Before(asOf)
}

// IsZero returns if the object is set or not.
// It will return true if either the userID or the sessionID are unset.
func (s *Session) IsZero() bool {
	return len(s.UserID) == 0 || len(s.SessionID) == 0
}
