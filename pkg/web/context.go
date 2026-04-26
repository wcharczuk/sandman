package web

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/text/language"
)

// Context is the struct that represents the context for an hc request.
type Context interface {
	context.Context
	// App is a reference back to the parent application.
	App() *App
	// Views should return the underlying views cache.
	Views() *Views
	// Response is the response writer for the request.
	Response() ResponseWriter
	// Request is the inbound request metadata.
	Request() *http.Request
	// Route is the matching route for the request if relevant.
	Route() *Route
	// RouteParams is a cache of parameters or variables
	// within the route and their values.
	RouteParams() RouteParameters
	// WithSession sets the session on the context.
	WithSession(*Session) Context
	// Session holds an active session if that is relevant.
	Session() *Session
	// Elapsed returns the current elapsed time for the request.
	Elapsed() time.Duration
	// Localization returns the i18n primitives.
	Localization() *Localization
	// Locale is the detected language preference.
	Locale() string
}

var (
	_ Context = (*baseContext)(nil)
)

type baseContext struct {
	app         *App
	req         *http.Request
	res         ResponseWriter
	route       *Route
	routeParams RouteParameters
	sess        *Session
	started     time.Time
}

func (bc *baseContext) App() *App                    { return bc.app }
func (bc *baseContext) Views() *Views                { return &bc.app.Views }
func (bc *baseContext) Response() ResponseWriter     { return bc.res }
func (bc *baseContext) Request() *http.Request       { return bc.req }
func (bc *baseContext) Route() *Route                { return bc.route }
func (bc *baseContext) RouteParams() RouteParameters { return bc.routeParams }
func (bc *baseContext) Session() *Session            { return bc.sess }
func (bc *baseContext) Elapsed() time.Duration       { return time.Since(bc.started) }
func (bc *baseContext) Localization() *Localization  { return &bc.app.Localization }

func (bc *baseContext) Locale() string {
	var locale string
	if queryLocale := bc.req.URL.Query().Get("locale"); queryLocale != "" {
		locale = queryLocale
	} else if headerAcceptLanguage := bc.req.Header.Get("Accept-Language"); headerAcceptLanguage != "" {
		tags, _, _ := language.ParseAcceptLanguage(headerAcceptLanguage)
		if len(tags) > 0 {
			locale = tags[0].String()
		}
	} else if session := bc.Session(); session != nil && session.Locale != "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en-us"
	}
	return locale
}

func (bc *baseContext) WithSession(sess *Session) Context {
	bc.sess = sess
	return bc
}

//
// context implementation
//

func (bc *baseContext) Deadline() (deadline time.Time, ok bool) {
	return bc.req.Context().Deadline()
}

func (bc *baseContext) Done() <-chan struct{} {
	return bc.req.Context().Done()
}

func (bc *baseContext) Err() error {
	return bc.req.Context().Err()
}

func (bc *baseContext) Value(key any) any {
	return bc.req.Context().Value(key)
}
