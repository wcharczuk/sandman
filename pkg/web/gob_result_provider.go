package web

import (
	"net/http"
)

var (
	// assert GobResultProbider implements ResultProvider.
	_ ResultProvider = (*GobResultProvider)(nil)
)

// GobResultProvider returns gob results for common responses.
type GobResultProvider struct{}

// NotFound returns a service response.
func (grp GobResultProvider) NotFound() Result {
	return &GobResult{
		StatusCode: http.StatusNotFound,
		Response:   http.StatusText(http.StatusNotFound),
	}
}

// NotAuthorized returns a service response.
func (grp GobResultProvider) NotAuthorized() Result {
	return &GobResult{
		StatusCode: http.StatusUnauthorized,
		Response:   http.StatusText(http.StatusUnauthorized),
	}
}

// Forbidden returns a service response.
func (grp GobResultProvider) Forbidden() Result {
	return &GobResult{
		StatusCode: http.StatusForbidden,
		Response:   http.StatusText(http.StatusForbidden),
	}
}

// InternalError returns a service response.
func (grp GobResultProvider) InternalError(err error) Result {
	return &GobResult{
		StatusCode: http.StatusInternalServerError,
		Response:   err.Error(),
	}
}

// BadRequest returns a service response.
func (grp GobResultProvider) BadRequest(err error) Result {
	return &GobResult{
		StatusCode: http.StatusBadRequest,
		Response:   err.Error(),
	}
}

// OK returns a service response.
func (grp GobResultProvider) OK() Result {
	return &GobResult{
		StatusCode: http.StatusOK,
		Response:   "OK!",
	}
}

// Result returns a service response.
func (grp GobResultProvider) Result(statusCode int, result any) Result {
	return &GobResult{
		StatusCode: statusCode,
		Response:   result,
	}
}
