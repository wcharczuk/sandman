package web

import "go.charczuk.com/sdk/errutil"

// JSONResult is a json result.
type JSONResult struct {
	StatusCode int
	Response   any
	Err        error
}

// Render renders the result
func (jr *JSONResult) Render(ctx Context) error {
	err := WriteJSON(ctx.Response(), jr.StatusCode, jr.Response)
	return errutil.Append(err, jr.Err)
}
