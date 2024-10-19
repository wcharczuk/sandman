package apputil

import (
	"fmt"
	"strings"
	"time"

	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/oauth"
	"go.charczuk.com/sdk/uuid"
	"go.charczuk.com/sdk/web"
)

// ApplyOAuthProfileToUser applies an oauth proflie.
func ApplyOAuthProfileToUser(u *User, p oauth.Profile) {
	u.Email = p.Email
	u.GivenName = p.GivenName
	u.FamilyName = p.FamilyName
	u.Locale = p.Locale
	u.PictureURL = p.PictureURL
}

// User is a user
type User struct {
	ID           uuid.UUID `db:"id,pk"`
	CreatedUTC   time.Time `db:"created_utc"`
	LastLoginUTC time.Time `db:"last_login_utc"`
	LastSeenUTC  time.Time `db:"last_seen_utc"`
	ProfileID    string    `db:"profile_id"`
	GivenName    string    `db:"given_name"`
	FamilyName   string    `db:"family_name"`
	PictureURL   string    `db:"picture_url"`
	Locale       string    `db:"locale"`
	Email        string    `db:"email"`
}

// TableName returns the mapped table name.
func (u User) TableName() string { return "users" }

// Username returns the user section of the email.
func (u User) Username() string {
	username, _, found := strings.Cut(u.Email, "@")
	if found {
		return username
	}
	return u.Email
}

// NewTestUser returns a test user.
func NewTestUser() User {
	return User{
		ID:           uuid.V4(),
		CreatedUTC:   time.Now().UTC().Add(-time.Hour),
		LastLoginUTC: time.Now().UTC().Add(-time.Minute),
		LastSeenUTC:  time.Now().UTC().Add(-time.Second),
		ProfileID:    uuid.V4().String(),
		GivenName:    uuid.V4().String(),
		FamilyName:   uuid.V4().String(),
		PictureURL:   fmt.Sprintf("https://example.com/%s.jpg", uuid.V4().String()),
		Locale:       "en-us",
		Email:        fmt.Sprintf("%s@example.com", uuid.V4().String()),
	}
}

var (
	_ db.TableNameProvider = (*Session)(nil)
)

// Session holds session information for a user.
type Session struct {
	SessionID   string    `db:"session_id,pk"`
	UserID      uuid.UUID `db:"user_id"`
	BaseURL     string    `db:"base_url"`
	CreatedUTC  time.Time `db:"created_utc"`
	ExpiresUTC  time.Time `db:"expires_utc"`
	LastSeenUTC time.Time `db:"last_seen_utc"`
	UserAgent   string    `db:"user_agent"`
	RemoteAddr  string    `db:"remote_addr"`
	Locale      string    `db:"locale"`
}

// TableName implements db.TableNameProvider.
func (s Session) TableName() string { return "sessions" }

// IsZero returns if the object is set or not.
// It will return true if either the userID or the sessionID are unset.
func (s *Session) IsZero() bool {
	return len(s.UserID) == 0 || len(s.SessionID) == 0
}

// ApplyTo sets a web session from the values on the type session.
func (s *Session) ApplyTo(webSession *web.Session) {
	webSession.SessionID = s.SessionID
	webSession.UserID = s.UserID.String()
	webSession.BaseURL = s.BaseURL
	webSession.CreatedUTC = s.CreatedUTC
	webSession.ExpiresUTC = s.ExpiresUTC
	webSession.UserAgent = s.UserAgent
	webSession.RemoteAddr = s.RemoteAddr
	webSession.Locale = s.Locale
}

// NewTestSession returns a test session.
func NewTestSession(user *User) Session {
	return Session{
		SessionID:   web.NewSessionID(),
		UserID:      user.ID,
		BaseURL:     "http://localhost",
		CreatedUTC:  time.Now().UTC(),
		ExpiresUTC:  time.Now().UTC().Add(24 * time.Hour),
		LastSeenUTC: time.Now().UTC().Add(time.Hour),
		UserAgent:   "go-test",
		RemoteAddr:  "127.0.0.1",
		Locale:      "en-us",
	}
}

// SessionState is the session state type.
type SessionState struct {
	User *User
}
