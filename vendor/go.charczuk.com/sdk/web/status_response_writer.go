package web

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
)

var (
	_ ResponseWriter      = (*StatusResponseWriter)(nil)
	_ http.ResponseWriter = (*StatusResponseWriter)(nil)
	_ http.Flusher        = (*StatusResponseWriter)(nil)
	_ io.Closer           = (*StatusResponseWriter)(nil)
)

// NewStatusResponseWriter creates a new response writer.
func NewStatusResponseWriter(w http.ResponseWriter) *StatusResponseWriter {
	if typed, ok := w.(*StatusResponseWriter); ok {
		return typed
	}
	if typed, ok := w.(ResponseWriter); ok {
		return &StatusResponseWriter{
			innerResponse: typed.InnerResponse(),
		}
	}
	return &StatusResponseWriter{
		innerResponse: w,
	}
}

// StatusResponseWriter a better response writer
type StatusResponseWriter struct {
	innerResponse http.ResponseWriter
	statusCode    int
	contentLength int
}

// Write writes the data to the response.
func (rw *StatusResponseWriter) Write(b []byte) (int, error) {
	bytesWritten, err := rw.innerResponse.Write(b)
	rw.contentLength += bytesWritten
	return bytesWritten, err
}

// Header accesses the response header collection.
func (rw *StatusResponseWriter) Header() http.Header {
	return rw.innerResponse.Header()
}

// Hijack wraps response writer's Hijack function.
func (rw *StatusResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.innerResponse.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("inner responseWriter doesn't support Hijacker interface")
	}
	return hijacker.Hijack()
}

// WriteHeader writes the status code (it is a somewhat poorly chosen method name from the standard library).
func (rw *StatusResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.innerResponse.WriteHeader(code)
}

// InnerResponse returns the backing writer.
func (rw *StatusResponseWriter) InnerResponse() http.ResponseWriter {
	return rw.innerResponse
}

// Flush calls flush on the inner response writer if it is supported.
func (rw *StatusResponseWriter) Flush() {
	if typed, ok := rw.innerResponse.(http.Flusher); ok {
		typed.Flush()
	}
}

// StatusCode returns the status code.
func (rw *StatusResponseWriter) StatusCode() int {
	return rw.statusCode
}

// ContentLength returns the content length
func (rw *StatusResponseWriter) ContentLength() int {
	return rw.contentLength
}

// Close calls close on the inner response if it supports it.
func (rw *StatusResponseWriter) Close() error {
	if typed, ok := rw.innerResponse.(io.Closer); ok {
		return typed.Close()
	}
	return nil
}
