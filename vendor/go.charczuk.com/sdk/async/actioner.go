package async

import "context"

// Actioner is a type that can be used as a tracked action.
type Actioner[T, V any] interface {
	Action(context.Context, T) (V, error)
}

// ActionerFunc is a function that implements action.
type ActionerFunc[T, V any] func(context.Context, T) (V, error)

// Action implements actioner for the function.
func (af ActionerFunc[T, V]) Action(ctx context.Context, args T) (V, error) {
	return af(ctx, args)
}

// NoopActioner is an actioner type that does nothing.
type NoopActioner[T, V any] struct{}

// Action implements actioner
func (n NoopActioner[T, V]) Action(_ context.Context, _ T) (out V, err error) {
	return
}
