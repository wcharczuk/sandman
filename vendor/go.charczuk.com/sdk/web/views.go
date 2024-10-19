package web

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"text/template"

	"go.charczuk.com/sdk/errutil"
)

const (
	// TemplateBadRequest is the default template name for bad request view results.
	TemplateBadRequest = "bad_request"
	// TemplateInternalError is the default template name for internal server error view results.
	TemplateInternalError = "error"
	// TemplateNotFound is the default template name for not found error view results.
	TemplateNotFound = "not_found"
	// TemplateNotAuthorized is the default template name for not authorized error view results.
	TemplateNotAuthorized = "not_authorized"
	// TemplateResult is the default template name for the result catchall endpoint.
	TemplateResult = "result"
)

var (
	// ErrUnsetViewTemplate is an error that is thrown if a given secure session id is invalid.
	ErrUnsetViewTemplate = errors.New("view result template is unset")
)

const (
	templateLiteralHeader        = `<html><head><style>body { font-family: sans-serif; text-align: center; }</style></head><body>`
	templateLiteralFooter        = `</body></html>`
	templateLiteralBadRequest    = templateLiteralHeader + `<h4>Bad Request</h4></body><pre>{{ . }}</pre>` + templateLiteralFooter
	templateLiteralInternalError = templateLiteralHeader + `<h4>Internal Error</h4><pre>{{ . }}</pre>` + templateLiteralFooter
	templateLiteralNotAuthorized = templateLiteralHeader + `<h4>Not Authorized</h4>` + templateLiteralFooter
	templateLiteralNotFound      = templateLiteralHeader + `<h4>Not Found</h4>` + templateLiteralFooter
	templateLiteralResult        = templateLiteralHeader + `<h4>{{ . }}</h4>` + templateLiteralFooter
)

var (
	_ ResultProvider = (*Views)(nil)
)

// Views is the cached views used in view results.
type Views struct {
	FuncMap template.FuncMap

	ViewPaths    []string
	ViewLiterals []string
	ViewFS       []ViewFS

	bp *BufferPool
	t  *template.Template
}

// ViewFS is a fs reference for views.
type ViewFS struct {
	// FS is the virtual filesystem reference.
	FS fs.FS
	// Patterns denotes glob patterns to match
	// within the filesystem itself (and can be empty!)
	Patterns []string
}

// AddPaths adds paths to the view collection.
func (vc *Views) AddPaths(paths ...string) {
	vc.ViewPaths = append(vc.ViewPaths, paths...)
}

// AddLiterals adds view literal strings to the view collection.
func (vc *Views) AddLiterals(views ...string) {
	vc.ViewLiterals = append(vc.ViewLiterals, views...)
}

// AddFS adds view fs instances to the view collection.
func (vc *Views) AddFS(fs ...ViewFS) {
	vc.ViewFS = append(vc.ViewFS, fs...)
}

// Initialize caches templates by path.
func (vc *Views) Initialize() (err error) {
	if vc.t == nil {
		if len(vc.ViewPaths) > 0 || len(vc.ViewLiterals) > 0 || len(vc.ViewFS) > 0 {
			vc.t, err = vc.Parse()
			if err != nil {
				err = fmt.Errorf("view initialize; %w", err)
				return
			}
		} else {
			vc.t = template.New("")
		}
	}
	if vc.bp == nil {
		vc.bp = NewBufferPool(256)
	}
	return
}

// Parse parses the view tree.
func (vc Views) Parse() (views *template.Template, err error) {
	views = template.New("views").Funcs(vc.FuncMap).Funcs(RequestFuncStubs())
	if len(vc.ViewPaths) > 0 {
		views, err = views.ParseFiles(vc.ViewPaths...)
		if err != nil {
			err = fmt.Errorf("cannot parse view files: %w", err)
			return
		}
	}
	for _, viewLiteral := range vc.ViewLiterals {
		views, err = views.Parse(viewLiteral)
		if err != nil {
			err = fmt.Errorf("cannot parse view literals: %w", err)
			return
		}
	}
	for _, viewFS := range vc.ViewFS {
		views, err = views.ParseFS(viewFS.FS, viewFS.Patterns...)
		if err != nil {
			err = fmt.Errorf("cannot parse view filesystems: %w", err)
			return
		}
	}
	return
}

// Lookup looks up a view.
func (vc Views) Lookup(name string) *template.Template {
	// we do the nil check here because there are usage patterns
	// where we may not have initialized the template cache
	// but still want to use Lookup.
	if vc.t == nil {
		return nil
	}
	return vc.t.Lookup(name)
}

// BadRequest returns a view result.
func (vc Views) BadRequest(err error) Result {
	t := vc.Lookup(TemplateBadRequest)
	if t == nil {
		t, _ = template.New("").Parse(templateLiteralBadRequest)
	}
	return &ViewResult{
		ViewName:   TemplateBadRequest,
		StatusCode: http.StatusBadRequest,
		ViewModel:  err,
		Template:   t,
		Views:      vc,
	}
}

// InternalError returns a view result.
func (vc Views) InternalError(err error) Result {
	t := vc.Lookup(TemplateInternalError)
	if t == nil {
		t, _ = template.New("").Parse(templateLiteralInternalError)
	}
	return &ViewResult{
		ViewName:   TemplateInternalError,
		StatusCode: http.StatusInternalServerError,
		ViewModel:  err,
		Template:   t,
		Views:      vc,
		Err:        err,
	}
}

// NotFound returns a view result.
func (vc Views) NotFound() Result {
	t := vc.Lookup(TemplateNotFound)
	if t == nil {
		t, _ = template.New("").Parse(templateLiteralNotFound)
	}
	return &ViewResult{
		ViewName:   TemplateNotFound,
		StatusCode: http.StatusNotFound,
		Template:   t,
		Views:      vc,
	}
}

// NotAuthorized returns a view result.
func (vc Views) NotAuthorized() Result {
	t := vc.Lookup(TemplateNotAuthorized)
	if t == nil {
		t, _ = template.New("").Parse(templateLiteralNotAuthorized)
	}
	return &ViewResult{
		ViewName:   TemplateNotAuthorized,
		StatusCode: http.StatusUnauthorized,
		Template:   t,
		Views:      vc,
	}
}

// Result returns a status view result.
func (vc Views) Result(statusCode int, response any) Result {
	t := vc.Lookup(TemplateResult)
	if t == nil {
		t, _ = template.New("").Parse(templateLiteralResult)
	}
	return &ViewResult{
		Views:      vc,
		ViewName:   TemplateResult,
		StatusCode: statusCode,
		Template:   t,
		ViewModel:  response,
	}
}

// View returns a view result with an OK status code.
func (vc Views) View(viewName string, viewModel any) Result {
	t := vc.Lookup(viewName)
	if t == nil {
		return vc.InternalError(ErrUnsetViewTemplate)
	}
	return &ViewResult{
		ViewName:   viewName,
		StatusCode: http.StatusOK,
		ViewModel:  viewModel,
		Template:   t,
		Views:      vc,
	}
}

// ViewStatus returns a view result with a given status code.
func (vc Views) ViewStatus(statusCode int, viewName string, viewModel any) Result {
	t := vc.Lookup(viewName)
	if t == nil {
		return vc.InternalError(ErrUnsetViewTemplate)
	}
	return &ViewResult{
		ViewName:   viewName,
		StatusCode: statusCode,
		ViewModel:  viewModel,
		Template:   t,
		Views:      vc,
	}
}

// ViewResult is a result that renders a view.
type ViewResult struct {
	ViewName   string
	StatusCode int
	ViewModel  any
	Views      Views
	Template   *template.Template
	Err        error
}

// Render renders the result to the given response writer.
func (vr *ViewResult) Render(ctx Context) (err error) {
	if vr.Template == nil {
		err = errutil.Append(ErrUnsetViewTemplate, vr.Err)
		return
	}
	ctx.Response().Header().Set(HeaderContentType, ContentTypeHTML)

	buffer := vr.Views.bp.Get()
	defer vr.Views.bp.Put(buffer)

	executeErr := vr.Template.Funcs(vr.RequestFuncs(ctx)).Execute(buffer, vr.ViewModel)
	if executeErr != nil {
		ctx.Response().WriteHeader(http.StatusInternalServerError)
		_, writeErr := ctx.Response().Write([]byte(fmt.Sprintf("%+v\n", executeErr)))
		err = errutil.Append(executeErr, writeErr, vr.Err)
		return
	}

	ctx.Response().WriteHeader(vr.StatusCode)
	_, writeErr := ctx.Response().Write(buffer.Bytes())
	err = errutil.Append(writeErr, vr.Err)
	return
}

// RequestFuncStubs are "stub" versions of the request bound funcs.
func RequestFuncStubs() template.FuncMap {
	return template.FuncMap{
		"request_context": func() Context { return nil },
		"localize":        func(str string) string { return str },
	}
}

// RequestFuncs returns the view funcs that are bound to the request specifically.
func (vr *ViewResult) RequestFuncs(ctx Context) template.FuncMap {
	return template.FuncMap{
		"request_context": func() Context { return ctx },
		"localize": func(str string) (string, error) {
			return ctx.App().Localization.Printer(ctx.Locale()).Print(str), nil
		},
	}
}
