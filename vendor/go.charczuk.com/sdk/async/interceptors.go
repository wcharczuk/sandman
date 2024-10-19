package async

// Interceptors chains calls to interceptors as a single interceptor.
func Interceptors[T, V any](interceptors ...Interceptor[T, V]) Interceptor[T, V] {
	if len(interceptors) == 0 {
		return nil
	}
	if len(interceptors) == 1 {
		return interceptors[0]
	}

	var curry = func(a, b Interceptor[T, V]) Interceptor[T, V] {
		if a == nil && b == nil {
			return nil
		}
		if a == nil {
			return b
		}
		if b == nil {
			return a
		}
		return InterceptorFunc[T, V](func(i Actioner[T, V]) Actioner[T, V] {
			return b.Intercept(a.Intercept(i))
		})
	}
	interceptor := interceptors[0]
	for _, next := range interceptors[1:] {
		interceptor = curry(interceptor, next)
	}
	return interceptor
}

// Interceptor returns an actioner for a given actioner.
type Interceptor[T, V any] interface {
	Intercept(action Actioner[T, V]) Actioner[T, V]
}

var (
	_ Interceptor[int, int] = (*InterceptorFunc[int, int])(nil)
)

// InterceptorFunc is a function that implements action.
type InterceptorFunc[T, V any] func(Actioner[T, V]) Actioner[T, V]

// Intercept implements Interceptor for the function.
func (fn InterceptorFunc[T, V]) Intercept(action Actioner[T, V]) Actioner[T, V] {
	return fn(action)
}
