package web

// GobResult is a gob rendered result.
type GobResult struct {
	StatusCode int
	Response   any
}

// Render renders the result
func (gr *GobResult) Render(ctx Context) error {
	return WriteGob(ctx.Response(), gr.StatusCode, gr.Response)
}
