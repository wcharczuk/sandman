package web

import (
	"net/http"
)

var (
	// assert it implements result provider.
	_ ResultProvider = (*JSONResultProvider)(nil)
)

// JSON returns a json result provider.
func JSON() JSONResultProvider {
	return JSONResultProvider{}
}

// JSONResultProvider are context results for api methods.
type JSONResultProvider struct{}

// NotFound returns a service response.
func (jrp JSONResultProvider) NotFound() Result {
	return &JSONResult{
		StatusCode: http.StatusNotFound,
		Response:   http.StatusText(http.StatusNotFound),
	}
}

// NotAuthorized returns a service response.
func (jrp JSONResultProvider) NotAuthorized() Result {
	return &JSONResult{
		StatusCode: http.StatusUnauthorized,
		Response:   http.StatusText(http.StatusUnauthorized),
	}
}

// Forbidden returns a service response.
func (jrp JSONResultProvider) Forbidden() Result {
	return &JSONResult{
		StatusCode: http.StatusForbidden,
		Response:   http.StatusText(http.StatusForbidden),
	}
}

// InternalError returns a json response.
func (jrp JSONResultProvider) InternalError(err error) Result {
	return &JSONResult{
		StatusCode: http.StatusInternalServerError,
		Response:   err.Error(),
		Err:        err,
	}
}

// BadRequest returns a json response.
func (jrp JSONResultProvider) BadRequest(err error) Result {
	return &JSONResult{
		StatusCode: http.StatusBadRequest,
		Response:   err.Error(),
	}
}

// OK returns a json response.
func (jrp JSONResultProvider) OK() Result {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response:   "OK!",
	}
}

// Result returns a json response.
func (jrp JSONResultProvider) Result(statusCode int, result any) Result {
	return &JSONResult{
		StatusCode: statusCode,
		Response:   result,
	}
}
