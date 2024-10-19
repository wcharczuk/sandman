package apputil

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/oauth"
	"go.charczuk.com/sdk/uuid"
	"go.charczuk.com/sdk/web"
)

// Auth is the auth controller.
type Auth struct {
	BaseController
	Config               Config
	AuthedRedirectPath   string
	CreateUserListener   func(context.Context, *User) error
	FetchSessionListener func(context.Context, *web.Session) error
	OAuth                *oauth.Manager
	DB                   *ModelManager
}

var (
	_ web.FetchSessionHandler   = (*Auth)(nil)
	_ web.PersistSessionHandler = (*Auth)(nil)
	_ web.RemoveSessionHandler  = (*Auth)(nil)
	_ web.LoginRedirectHandler  = (*Auth)(nil)
)

// Register adds the controller routes to the application.
func (a Auth) Register(app *web.App) {
	app.AuthPersister = a

	app.Get("/login", web.SessionAware(a.login))
	app.Get("/logout", web.SessionAwareStable(a.logout))
	app.Get("/oauth/google", web.SessionAware(a.oauthGoogle))
}

// GET /login
func (a Auth) login(r web.Context) web.Result {
	if r.Session() != nil {
		return a.authedRedirect()
	}
	oauthURL, err := a.OAuth.OAuthURL(r.Request(), oauth.OptStateRedirectURI(r.Request().RequestURI))
	if err != nil {
		return r.Views().InternalError(err)
	}
	return web.RedirectWithMethod("GET", oauthURL)
}

// GET /oauth/google
func (a Auth) oauthGoogle(ctx web.Context) web.Result {
	if ctx.Session() != nil {
		return a.authedRedirect()
	}
	result, err := a.OAuth.Finish(ctx.Request())
	if err != nil {
		log.GetLogger(ctx).WithGroup("web").Error("authentication error", err)
		return ctx.App().Views.NotAuthorized()
	}

	user, existingUserFound, err := a.DB.GetUserByEmail(ctx, result.Profile.Email)
	if err != nil {
		return ctx.App().Views.InternalError(err)
	}
	ApplyOAuthProfileToUser(&user, result.Profile)
	if !existingUserFound {
		user.ID = uuid.V4()
		user.CreatedUTC = time.Now().UTC()
	}
	user.LastLoginUTC = time.Now().UTC()
	user.LastSeenUTC = time.Now().UTC()

	if err = a.DB.Invoke(ctx).Upsert(&user); err != nil {
		return ctx.Views().InternalError(err)
	}

	if !existingUserFound && a.CreateUserListener != nil {
		if err = a.CreateUserListener(ctx, &user); err != nil {
			return ctx.Views().InternalError(err)
		}
	}

	_, err = ctx.App().Login(user.ID.String(), ctx)
	if err != nil {
		return ctx.Views().InternalError(err)
	}
	if len(result.State.RedirectURI) > 0 {
		return web.RedirectWithMethodf(http.MethodGet, result.State.RedirectURI)
	}
	return a.authedRedirect()
}

// logout logs the user out.
func (a Auth) logout(ctx web.Context) web.Result {
	if ctx.Session() == nil {
		return ctx.App().Views.NotAuthorized()
	}
	if err := ctx.App().Logout(ctx); err != nil {
		return ctx.Views().InternalError(err)
	}
	return web.RedirectWithMethod("GET", "/")
}

//
// helpers
//

func (a Auth) authedRedirect() web.Result {
	redirectTargetPath := "/"
	if a.AuthedRedirectPath != "" {
		redirectTargetPath = a.AuthedRedirectPath
	}
	return web.RedirectWithMethod(http.MethodGet, redirectTargetPath)
}

// SessionTimeout implements session timeout provider.
func (a Auth) SessionTimeout(_ *web.Session) time.Time {
	return time.Now().UTC().AddDate(0, 0, 14)
}

// FetchSession implements web.FetchSessionHandler.
func (a Auth) FetchSession(ctx context.Context, sessionID string) (*web.Session, error) {
	var dbSession Session
	_, err := a.DB.Invoke(ctx).Get(&dbSession, sessionID)
	if err != nil {
		return nil, err
	}
	if dbSession.IsZero() {
		return nil, nil
	}

	var session web.Session
	dbSession.ApplyTo(&session)
	var user User
	if _, err = a.DB.Invoke(ctx).Get(&user, dbSession.UserID); err != nil {
		return nil, err
	}
	session.State = SessionState{
		User: &user,
	}
	if a.FetchSessionListener != nil {
		if err = a.FetchSessionListener(ctx, &session); err != nil {
			return nil, err
		}
	}
	return &session, nil
}

// PersistSession implements web.PersistSessionHandler.
func (a Auth) PersistSession(ctx context.Context, session *web.Session) error {
	dbSession := &Session{
		SessionID:   session.SessionID,
		UserID:      uuid.MustParse(session.UserID),
		BaseURL:     session.BaseURL,
		CreatedUTC:  session.CreatedUTC,
		ExpiresUTC:  session.ExpiresUTC,
		LastSeenUTC: time.Now().UTC(),
		UserAgent:   session.UserAgent,
		RemoteAddr:  session.RemoteAddr,
		Locale:      session.Locale,
	}
	return a.DB.Invoke(ctx).Upsert(dbSession)
}

// RemoveSession implements web.RemoveSessionHandler.
func (a Auth) RemoveSession(ctx context.Context, sessionID string) error {
	var session Session
	_, err := a.DB.Invoke(ctx).Get(&session, sessionID)
	if err != nil {
		return err
	}
	_, err = a.DB.Invoke(ctx).Delete(session)
	return err
}

// LoginRedirect implements web.LoginRedirectHandler.
func (a Auth) LoginRedirect(ctx web.Context) *url.URL {
	from := ctx.Request().URL.Path
	oauthURL, err := a.OAuth.OAuthURL(ctx.Request(), oauth.OptStateRedirectURI(from))
	if err != nil {
		return &url.URL{RawPath: "/login?error=invalid_oauth_url"}
	}
	parsed, err := url.Parse(oauthURL)
	if err != nil {
		return &url.URL{RawPath: "/login?error=invalid_oauth_url"}
	}
	return parsed
}
