package web

// Action is the function signature for controller actions.
type Action func(Context) Result

// PanicAction is a receiver for app.PanicHandler.
type PanicAction func(Context, interface{}) Result
