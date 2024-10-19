package web

import (
	"context"
	"net/http"
	"time"

	"go.charczuk.com/sdk/errutil"
	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/viewutil"
)

// New returns a new app with defaults configured and applies
// a given set of options to it.
//
// You can alternately use the `app := new(App)` form but it's
// highly recommended to use this constructor.
func New() *App {
	app := new(App)
	app.Views.FuncMap = viewutil.Funcs
	app.NotFoundHandler = app.ActionHandler(defaultNotFoundHandler)
	app.PanicAction = defaultPanicAction
	return app
}

func defaultNotFoundHandler(ctx Context) Result {
	return AcceptedProvider(ctx).NotFound()
}

func defaultPanicAction(ctx Context, r any) Result {
	return AcceptedProvider(ctx).InternalError(errutil.New(r))
}

// App is the container type for the application and its resources.
//
// Typical usage is just to instantiate it with a bare constructor:
//
//	app := new(web.App)
//	app.Get("/", func(ctx web.Context) web.Result {
//		return &web.JSONResult{StatusCode: http.StatusOK, Response: "OK!" }
//	}
type App struct {
	Auth
	RouteTree
	Server

	Views        Views
	Localization Localization

	BaseURL     string
	Headers     http.Header
	PanicAction PanicAction

	middleware []Middleware
	onRequest  []func(RequestEvent)
	onError    []func(Context, error)
}

// RegisterControllers registers controllers.
func (a *App) RegisterControllers(controllers ...Controller) {
	for _, c := range controllers {
		c.Register(a)
	}
}

// RegisterMiddleware registers global middleware.
func (a *App) RegisterMiddleware(middleware ...Middleware) {
	a.middleware = append(a.middleware, middleware...)
}

// RegisterConfig registers the config.
func (a *App) RegisterConfig(cfg Config) {
	cfg.ApplyTo(a)
}

// RegisterLoggerListeners registers a logger with the listen,
// request, and error handler lists.
func (a *App) RegisterLoggerListeners(log *log.Logger) {
	a.RegisterOnListen(LogOnListen(a, log))
	a.RegisterOnRequest(LogOnRequest(log))
	a.RegisterOnError(LogOnError(log))
}

// RegisterOnRequest adds an on request hook.
func (a *App) RegisterOnRequest(fn func(RequestEvent)) {
	a.onRequest = append(a.onRequest, fn)
}

// RegisterOnError adds an on request hook.
func (a *App) RegisterOnError(fn func(Context, error)) {
	a.onError = append(a.onError, fn)
}

// Get registers a GET request route handler.
//
// To add additional middleware, use  `web.NextMiddleware(action, ...middleware)`.
func (a *App) Get(path string, action Action) {
	a.HandleAction(http.MethodGet, path, action)
}

// Options registers a OPTIONS request route handler.
//
// To add additional middleware, use  `web.NextMiddleware(action, ...middleware)`.
func (a *App) Options(path string, action Action) {
	a.HandleAction(http.MethodOptions, path, action)
}

// Head registers a HEAD request route handler.
//
// To add additional middleware, use  `web.NextMiddleware(action, ...middleware)`.
func (a *App) Head(path string, action Action) {
	a.HandleAction(http.MethodHead, path, action)
}

// Put registers a PUT request route handler.
//
// To add additional middleware, use `web.NextMiddleware(action, ...middleware)`.
func (a *App) Put(path string, action Action) {
	a.HandleAction(http.MethodPut, path, action)
}

// Patch registers a PATCH request route handler.
//
// To add additional middleware, use `web.NextMiddleware(action, ...middleware)`.
func (a *App) Patch(path string, action Action) {
	a.HandleAction(http.MethodPatch, path, action)
}

// Post registers a POST request route handler.
//
// To add additional middleware, use  `web.NextMiddleware(action, ...middleware)`.
func (a *App) Post(path string, action Action) {
	a.HandleAction(http.MethodPost, path, action)
}

// Delete registers a DELETE request route handler.
//
// To add additional middleware, use  `web.NextMiddleware(action, ...middleware)`.
func (a *App) Delete(path string, action Action) {
	a.HandleAction(http.MethodDelete, path, action)
}

// HandleAction registers an action for a given method and path with the given middleware.
func (a *App) HandleAction(method string, path string, action Action) {
	a.Handle(method, path, a.ActionHandler(NestMiddleware(action, a.middleware...)))
}

// Lookup finds the route data for a given method and path.
func (a *App) Lookup(method, path string) (route *Route, params RouteParameters, skipSlashRedirect bool) {
	if root := a.Routes[method]; root != nil {
		route, params, skipSlashRedirect = root.getPath(path)
		return
	}
	return
}

// ActionHandler is the translation step from Action to Handler.
func (a *App) ActionHandler(action Action) Handler {
	return func(w http.ResponseWriter, r *http.Request, route *Route, p RouteParameters) {
		ctx := &baseContext{
			app:         a,
			req:         r,
			res:         NewStatusResponseWriter(w),
			route:       route,
			routeParams: p,
			started:     time.Now().UTC(),
		}
		if a.PanicAction != nil {
			defer func() {
				if r := recover(); r != nil {
					if res := a.PanicAction(ctx, r); res != nil {
						a.renderResult(ctx, res)
					}
					a.handleError(ctx, errutil.New(r))
				}
			}()
		}
		for key, value := range a.Headers {
			ctx.Response().Header()[key] = value
		}
		if res := action(ctx); res != nil {
			a.renderResult(ctx, res)
		}
		if len(a.onRequest) > 0 {
			re := NewRequestEvent(ctx)
			for _, reqHandler := range a.onRequest {
				reqHandler(re)
			}
		}
	}
}

// Initialize calls initializers on the localization engine
// and the views cache.
//
// If you don't use either of these facilities (e.g. if you're only
// implementing JSON api routes, you can skip calling this function).
func (a *App) Initialize() error {
	if err := a.Localization.Initialize(); err != nil {
		return err
	}
	if err := a.Views.Initialize(); err != nil {
		return err
	}
	return nil
}

// Start implements the first phase of graceful.Graceful
// and initializes the view cache and does some other housekeeping
// before starting the server.
func (a *App) Start(ctx context.Context) error {
	if err := a.Views.Initialize(); err != nil {
		return err
	}
	a.Handler = a.RouteTree
	return a.Server.Start(ctx)
}

//
// internal helpers
//

func (a *App) renderResult(ctx Context, res Result) {
	if err := res.Render(ctx); err != nil {
		a.handleError(ctx, err)
	}
}

func (a *App) handleError(ctx Context, err error) {
	if err == nil {
		return
	}
	for _, errHandler := range a.onError {
		errHandler(ctx, err)
	}
}
