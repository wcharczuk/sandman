package web

// Middleware is a func that implements middleware
type Middleware func(Action) Action
