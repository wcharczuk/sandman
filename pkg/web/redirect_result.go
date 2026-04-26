package web

import (
	"fmt"
	"net/http"
)

// Redirect returns a redirect result to a given destination.
func Redirect(destination string) *RedirectResult {
	return &RedirectResult{
		RedirectURI: destination,
	}
}

// RedirectStatus returns a redirect result to a given destination with a given status code.
func RedirectStatus(destination string, statusCode int) *RedirectResult {
	return &RedirectResult{
		RedirectURI: destination,
		StatusCode:  statusCode,
	}
}

// Redirectf returns a redirect result to a given destination specified by a given format and scan arguments.
func Redirectf(format string, args ...interface{}) *RedirectResult {
	return &RedirectResult{
		RedirectURI: fmt.Sprintf(format, args...),
	}
}

// RedirectWithMethod returns a redirect result to a destination with a given method.
func RedirectWithMethod(method, destination string) *RedirectResult {
	return &RedirectResult{
		Method:      method,
		RedirectURI: destination,
	}
}

// RedirectWithMethodf returns a redirect result to a destination composed of a format and scan arguments with a given method.
func RedirectWithMethodf(method, format string, args ...interface{}) *RedirectResult {
	return &RedirectResult{
		Method:      method,
		RedirectURI: fmt.Sprintf(format, args...),
	}
}

// RedirectResult is a result that should cause the browser to redirect.
type RedirectResult struct {
	Method      string `json:"redirect_method"`
	RedirectURI string `json:"redirect_uri"`
	StatusCode  int    `json:"status_code"`
}

// Render writes the result to the response.
func (rr *RedirectResult) Render(ctx Context) error {
	if len(rr.Method) > 0 {
		ctx.Request().Method = rr.Method
		statusCode := http.StatusFound
		if rr.StatusCode > 0 {
			statusCode = rr.StatusCode
		}
		http.Redirect(ctx.Response(), ctx.Request(), rr.RedirectURI, statusCode)
		return nil
	}
	statusCode := http.StatusTemporaryRedirect
	if rr.StatusCode > 0 {
		statusCode = rr.StatusCode
	}
	http.Redirect(ctx.Response(), ctx.Request(), rr.RedirectURI, statusCode)
	return nil
}
