package web

// ResultProvider is the provider interface for status based results.
type ResultProvider interface {
	InternalError(error) Result
	BadRequest(error) Result
	NotFound() Result
	NotAuthorized() Result
	Result(int, any) Result
}
