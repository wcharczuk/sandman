package web

// Result is the type returned by actions.
type Result interface {
	Render(Context) error
}
