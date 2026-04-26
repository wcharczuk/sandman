package web

// Controller is a type that can be Registered.
type Controller interface {
	Register(*App)
}
