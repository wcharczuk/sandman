package web

import (
	"compress/gzip"
	"net/http"
)

// GZip is a middleware the implements gzip compression for requests that opt into it.
func GZip(action Action) Action {
	return func(r Context) Result {
		if HeaderAny(r.Request().Header, HeaderAcceptEncoding, ContentEncodingGZIP) {
			r.Response().Header().Set(HeaderContentEncoding, ContentEncodingGZIP)
			r.Response().Header().Set(HeaderVary, HeaderAcceptEncoding)
			return action(&baseContext{
				app:         r.App(),
				req:         r.Request(),
				route:       r.Route(),
				routeParams: r.RouteParams(),
				res:         NewGZipResponseWriter(r.Response()),
			})
		}
		return action(r)
	}
}

var (
	_ ResponseWriter      = (*GZipResponseWriter)(nil)
	_ http.ResponseWriter = (*GZipResponseWriter)(nil)
	_ http.Flusher        = (*GZipResponseWriter)(nil)
)

// NewGZipResponseWriter returns a new gzipped response writer.
func NewGZipResponseWriter(w http.ResponseWriter) *GZipResponseWriter {
	if typed, ok := w.(ResponseWriter); ok {
		return &GZipResponseWriter{
			innerResponse: typed.InnerResponse(),
			gzipWriter:    gzip.NewWriter(typed.InnerResponse()),
		}
	}
	return &GZipResponseWriter{
		innerResponse: w,
		gzipWriter:    gzip.NewWriter(w),
	}
}

// GZipResponseWriter is a response writer that compresses output.
type GZipResponseWriter struct {
	gzipWriter    *gzip.Writer
	innerResponse http.ResponseWriter
	statusCode    int
	contentLength int
}

// InnerResponse returns the underlying response.
func (crw *GZipResponseWriter) InnerResponse() http.ResponseWriter {
	return crw.innerResponse
}

// Write writes the byes to the stream.
func (crw *GZipResponseWriter) Write(b []byte) (int, error) {
	_, err := crw.gzipWriter.Write(b)
	crw.contentLength += len(b)
	return len(b), err
}

// Header returns the headers for the response.
func (crw *GZipResponseWriter) Header() http.Header {
	return crw.innerResponse.Header()
}

// WriteHeader writes a status code.
func (crw *GZipResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.innerResponse.WriteHeader(code)
}

// StatusCode returns the status code for the request.
func (crw *GZipResponseWriter) StatusCode() int {
	return crw.statusCode
}

// ContentLength returns the content length for the request.
func (crw *GZipResponseWriter) ContentLength() int {
	return crw.contentLength
}

// Flush pushes any buffered data out to the response.
func (crw *GZipResponseWriter) Flush() {
	crw.gzipWriter.Flush()
}

// Close closes any underlying resources.
func (crw *GZipResponseWriter) Close() error {
	return crw.gzipWriter.Close()
}
