package async

import "context"

// Action is a type that can be used in action context chains.
type Action[T, V any] interface {
	Call(context.Context, T) (V, error)
}

var _ Action[string, int] = (*ActionFunc[string, int])(nil)

// ActionFunc is a function that implements action.
type ActionFunc[T, V any] func(context.Context, T) (V, error)

// Action implements actioner for the function.
func (af ActionFunc[T, V]) Call(ctx context.Context, args T) (V, error) {
	return af(ctx, args)
}

var _ Action[string, int] = (*NoopAction[string, int])(nil)

// NoopAction is an actioner type that does nothing.
type NoopAction[T, V any] struct{}

// Action implements actioner
func (n NoopAction[T, V]) Call(_ context.Context, _ T) (out V, err error) {
	return
}
