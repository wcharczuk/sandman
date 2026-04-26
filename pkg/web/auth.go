package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"time"
)

// NewSessionID returns a new session id.
// It is not a uuid; session ids are generated using a secure random source.
// SessionIDs are generally 64 bytes.
func NewSessionID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

var (
	// ErrSessionIDEmpty is thrown if a session id is empty.
	ErrSessionIDEmpty = errors.New("auth session id is empty")
	// ErrSecureSessionIDEmpty is an error that is thrown if a given secure session id is invalid.
	ErrSecureSessionIDEmpty = errors.New("auth secure session id is empty")
)

// IsErrSessionInvalid returns if an error is a session invalid error.
func IsErrSessionInvalid(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrSessionIDEmpty) || errors.Is(err, ErrSecureSessionIDEmpty) {
		return true
	}
	return false
}

const (
	// DefaultCookieName is the default name of the field that contains the session id.
	DefaultCookieName = "SID"
	// DefaultCookiePath is the default cookie path.
	DefaultCookiePath = "/"
	// DefaultCookieSecure returns what the default value for the `Secure` bit of issued cookies will be.
	DefaultCookieSecure = true
	// DefaultCookieHTTPOnly returns what the default value for the `HTTPOnly` bit of issued cookies will be.
	DefaultCookieHTTPOnly = true
	// DefaultCookieSameSiteMode is the default cookie same site mode (currently http.SameSiteLaxMode).
	DefaultCookieSameSiteMode = http.SameSiteLaxMode
	// DefaultSessionTimeout is the default absolute timeout for a session (24 hours as a sane default).
	DefaultSessionTimeout time.Duration = 24 * time.Hour
	// DefaultUseSessionCache is the default if we should use the auth manager session cache.
	DefaultUseSessionCache = true
	// DefaultSessionTimeoutIsAbsolute is the default if we should set absolute session expiries.
	DefaultSessionTimeoutIsAbsolute = true
)

// SerializeSessionHandler serializes a session as a string.
type SerializeSessionHandler interface {
	SerializeSession(context.Context, *Session) (string, error)
}

// PersistSessionHandler saves the session to a stable store.
type PersistSessionHandler interface {
	PersistSession(context.Context, *Session) error
}

// FetchSessionHandler restores a session based on a session value.
type FetchSessionHandler interface {
	FetchSession(context.Context, string) (*Session, error)
}

// RemoveSessionHandler removes a session based on a session value.
type RemoveSessionHandler interface {
	RemoveSession(context.Context, string) error
}

// ValidateSessionHandler validates a session.
type ValidateSessionHandler interface {
	ValidateSession(context.Context, *Session) error
}

// SessionTimeoutProvider provides a new timeout for a session.
type SessionTimeoutProvider interface {
	SessionTimeout(*Session) time.Time
}

// LoginRedirectHandler is a redirect handler.
type LoginRedirectHandler interface {
	LoginRedirect(Context) *url.URL
}

// Auth is a manager for sessions.
type Auth struct {
	// AuthCookieDefaults hold the default cookie options when creating cookies.
	AuthCookieDefaults http.Cookie
	// AuthPersister is the underlying auth persistence handler.
	AuthPersister any
}

// --------------------------------------------------------------------------------
// Methods
// --------------------------------------------------------------------------------

// Login logs a userID in.
func (a Auth) Login(userID string, ctx Context) (session *Session, err error) {
	// create a new session id (which will be a cookie value)
	sessionValue := NewSessionID()

	// userID and sessionID are required
	session = &Session{
		UserID:     userID,
		SessionID:  sessionValue,
		CreatedUTC: time.Now().UTC(),
	}

	if handler, ok := a.AuthPersister.(SessionTimeoutProvider); ok {
		session.ExpiresUTC = handler.SessionTimeout(session)
	}
	session.UserAgent = ctx.Request().UserAgent()
	session.RemoteAddr = GetRemoteAddr(ctx.Request())

	// call the perist handler if one's been provided
	if handler, ok := a.AuthPersister.(PersistSessionHandler); ok {
		err = handler.PersistSession(ctx, session)
		if err != nil {
			return nil, err
		}
	}

	// call the serialize handler if one's been provided
	if handler, ok := a.AuthPersister.(SerializeSessionHandler); ok {
		sessionValue, err = handler.SerializeSession(ctx, session)
		if err != nil {
			return nil, err
		}
	}

	// inject cookies into the response
	a.injectCookie(ctx, sessionValue, session.ExpiresUTC)
	return session, nil
}

// Logout unauthenticates a session.
func (a Auth) Logout(ctx Context) error {
	sessionValue := a.readSessionValue(ctx)
	// validate the sessionValue isn't unset
	if sessionValue == "" {
		return nil
	}

	// zero out the context session as a precaution
	// note: this is skipped ... why?
	// ctx.Session = nil

	// issue the expiration cookies to the response
	// and call the remove handler
	return a.expire(ctx, sessionValue)
}

// VerifySession pulls the session cookie off the request, and validates
// it represents a valid session.
func (a Auth) VerifySession(ctx Context) (sessionValue string, session *Session, err error) {
	sessionValue = a.readSessionValue(ctx)
	// validate the sessionValue is set
	if len(sessionValue) == 0 {
		return
	}

	// if we have a restore handler, call it.
	if handler, ok := a.AuthPersister.(FetchSessionHandler); ok {
		session, err = handler.FetchSession(ctx, sessionValue)
		if err != nil {
			if IsErrSessionInvalid(err) {
				_ = a.expire(ctx, sessionValue)
			}
			return
		}
	}

	// if the session is invalid, expire the cookie(s)
	if session == nil || session.IsZero() || session.IsExpired(time.Now().UTC()) {
		session = nil
		err = a.expire(ctx, sessionValue)
		return
	}

	// call a custom validate handler if one's been provided.
	if handler, ok := a.AuthPersister.(ValidateSessionHandler); ok {
		err = handler.ValidateSession(ctx, session)
		if err != nil {
			session = nil
			return
		}
	}
	return
}

// VerifyOrExtendSession reads a session value from a request and checks if it's valid.
// It also handles updating a rolling expiry.
func (a Auth) VerifyOrExtendSession(ctx Context) (session *Session, err error) {
	var sessionValue string
	sessionValue, session, err = a.VerifySession(ctx)
	if session == nil || err != nil {
		return
	}

	if handler, ok := a.AuthPersister.(SessionTimeoutProvider); ok {
		existingExpiresUTC := session.ExpiresUTC
		session.ExpiresUTC = handler.SessionTimeout(session)

		// if session expiry has changed
		if existingExpiresUTC != session.ExpiresUTC {
			// if we have a persist handler
			// call it to reflect the updated session timeout.
			if persistHandler, ok := a.AuthPersister.(PersistSessionHandler); ok {
				err = persistHandler.PersistSession(ctx, session)
				if err != nil {
					return nil, err
				}
			}
			// inject the (updated) cookie
			a.injectCookie(ctx, sessionValue, session.ExpiresUTC)
		}
	}
	return
}

// LoginRedirect returns a redirect result for when auth fails and you need to
// send the user to a login page.
func (a Auth) LoginRedirect(ctx Context) Result {
	if handler, ok := a.AuthPersister.(LoginRedirectHandler); ok {
		redirectTo := handler.LoginRedirect(ctx)
		if redirectTo != nil {
			return Redirect(redirectTo.String())
		}
	}
	return AcceptedProvider(ctx).NotAuthorized()
}

//
// Cookie Defaults
//

// CookieNameOrDefault returns the configured cookie name or a default.
func (a Auth) CookieNameOrDefault() string {
	if a.AuthCookieDefaults.Name == "" {
		return DefaultCookieName
	}
	return a.AuthCookieDefaults.Name
}

// CookiePathOrDefault returns the configured cookie path or a default.
func (a Auth) CookiePathOrDefault() string {
	if a.AuthCookieDefaults.Path == "" {
		return DefaultCookiePath
	}
	return a.AuthCookieDefaults.Path
}

// --------------------------------------------------------------------------------
// Utility Methods
// --------------------------------------------------------------------------------

func (a Auth) expire(ctx Context, sessionValue string) error {
	// issue the cookie expiration.
	a.expireCookie(ctx)

	// if we have a remove handler and the sessionID is set
	if handler, ok := a.AuthPersister.(RemoveSessionHandler); ok {
		err := handler.RemoveSession(ctx, sessionValue)
		if err != nil {
			return err
		}
	}
	return nil
}

// InjectCookie injects a session cookie into the context.
func (a Auth) injectCookie(ctx Context, value string, expire time.Time) {
	http.SetCookie(ctx.Response(), &http.Cookie{
		Value:    value,
		Expires:  expire,
		Name:     a.CookieNameOrDefault(),
		Path:     a.CookiePathOrDefault(),
		Domain:   a.AuthCookieDefaults.Domain,
		HttpOnly: a.AuthCookieDefaults.HttpOnly,
		Secure:   a.AuthCookieDefaults.Secure,
		SameSite: a.AuthCookieDefaults.SameSite,
	})
}

// expireCookie expires the session cookie.
func (a Auth) expireCookie(ctx Context) {
	// cookie MaxAge<0 means delete cookie now, and is equivalent to
	// the literal cookie header content 'Max-Age: 0'
	http.SetCookie(ctx.Response(), &http.Cookie{
		Value:    NewSessionID(),
		MaxAge:   -1,
		Name:     a.CookieNameOrDefault(),
		Path:     a.CookiePathOrDefault(),
		Domain:   a.AuthCookieDefaults.Domain,
		HttpOnly: a.AuthCookieDefaults.HttpOnly,
		Secure:   a.AuthCookieDefaults.Secure,
		SameSite: a.AuthCookieDefaults.SameSite,
	})
}

// cookieValue reads a param from a given request context from either the cookies or headers.
func (a Auth) cookieValue(name string, ctx Context) (output string) {
	if cookie, _ := ctx.Request().Cookie(name); cookie != nil {
		output = cookie.Value
	}
	return
}

// ReadSessionID reads a session id from a given request context.
func (a Auth) readSessionValue(ctx Context) string {
	return a.cookieValue(a.CookieNameOrDefault(), ctx)
}
