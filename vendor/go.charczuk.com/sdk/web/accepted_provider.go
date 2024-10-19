package web

// AcceptedProvider returns a result provider by the Accept
// header value on the incoming request.
func AcceptedProvider(ctx Context) ResultProvider {
	contentType, _ := NegotiateContentType(ctx.Request(),
		"application/json",
		"application/gob",
		"text/html",
		"text/plain",
	)

	switch contentType {
	case "application/json":
		return JSONResultProvider{}
	case ContentTypeApplicationGob:
		return GobResultProvider{}
	case "text/html":
		return ctx.Views()
	case "text/plain":
		return TextResultProvider{}
	default:
		return JSONResultProvider{}
	}
}
