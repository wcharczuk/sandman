package web

// NestMiddleware reads the middleware variadic args and organizes the calls
// recursively in the order they appear. I.e. NestMiddleware(inner, third,
// second, first) will call "first", "second", "third", then "inner".
func NestMiddleware(action Action, middleware ...Middleware) Action {
	if len(middleware) == 0 {
		return action
	}

	a := action
	for _, i := range middleware {
		if i != nil {
			a = i(a)
		}
	}
	return a
}
