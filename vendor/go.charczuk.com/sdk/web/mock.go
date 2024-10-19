package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"go.charczuk.com/sdk/r2"
)

// Mock sends a mock request to an app.
//
// It will reset the app Server, Listener, and will set the request host to the listener address
// for a randomized local listener.
func Mock(app *App, req *http.Request, options ...r2.Option) *MockResult {
	r2req := new(r2.Request)
	_ = r2.OptRequest(req)(r2req)

	var err error
	result := &MockResult{
		App:     app,
		Request: r2req,
	}
	for _, option := range options {
		if err = option(result.Request); err != nil {
			_ = r2.OptErr(err)(result.Request)
			return result
		}
	}
	if err := app.Initialize(); err != nil {
		_ = r2.OptErr(err)(result.Request)
		return result
	}

	if result.URL() == nil {
		_ = r2.OptURL(new(url.URL))(result.Request)
	}

	result.Server = httptest.NewUnstartedServer(app)
	// result.Server.Config.BaseContext = app.BaseContext
	result.Server.Start()
	_ = r2.OptCloser(result.Close)(result.Request)

	parsedServerURL := MustParseURL(result.Server.URL)
	_ = r2.OptScheme(parsedServerURL.Scheme)(result.Request)
	_ = r2.OptHost(parsedServerURL.Host)(result.Request)
	return result
}

// MockMethod sends a mock request with a given method to an app.
// You should use request options to set the body of the request if it's a post or put etc.
func MockMethod(app *App, method, path string, options ...r2.Option) *MockResult {
	req := &http.Request{
		Method: method,
		URL: &url.URL{
			Path: path,
		},
	}
	return Mock(app, req, options...)
}

// MockGet sends a mock get request to an app.
func MockGet(app *App, path string, options ...r2.Option) *MockResult {
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: path,
		},
	}
	return Mock(app, req, options...)
}

// MockPost sends a mock post request to an app.
func MockPost(app *App, path string, body io.ReadCloser, options ...r2.Option) *MockResult {
	req := &http.Request{
		Method: "POST",
		Body:   body,
		URL: &url.URL{
			Path: path,
		},
	}
	return Mock(app, req, options...)
}

// MockPostJSON sends a mock post request with a json body to an app.
func MockPostJSON(app *App, path string, body interface{}, options ...r2.Option) *MockResult {
	contents, _ := json.Marshal(body)
	req := &http.Request{
		Method: "POST",
		Body:   io.NopCloser(bytes.NewReader(contents)),
		URL: &url.URL{
			Path: path,
		},
	}
	return Mock(app, req, options...)
}

// MockResult is a result of a mocked request.
type MockResult struct {
	*r2.Request
	App    *App
	Server *httptest.Server
}

// Close stops the app.
func (mr *MockResult) Close() error {
	mr.Server.Close()
	return nil
}

// MockContext returns a new mock ctx.
// It is intended to be used in testing.
func MockContext(method, path string) Context {
	return MockContextWithBuffer(method, path, new(bytes.Buffer))
}

// MockContextWithBuffer returns a new mock ctx.
// It is intended to be used in testing.
func MockContextWithBuffer(method, path string, buf io.Writer) Context {
	return &baseContext{
		app: new(App),
		res: NewMockResponse(buf),
		req: NewMockRequest(method, path),
	}
}

var (
	_ http.ResponseWriter = (*MockResponseWriter)(nil)
	_ http.Flusher        = (*MockResponseWriter)(nil)
)

// NewMockResponse returns a mocked response writer.
func NewMockResponse(buffer io.Writer) *MockResponseWriter {
	return &MockResponseWriter{
		innerWriter: buffer,
		contents:    new(bytes.Buffer),
		headers:     http.Header{},
	}
}

// MockResponseWriter is an object that satisfies response writer but uses an internal buffer.
type MockResponseWriter struct {
	innerWriter   io.Writer
	contents      *bytes.Buffer
	statusCode    int
	contentLength int
	headers       http.Header
}

// Write writes data and adds to ContentLength.
func (res *MockResponseWriter) Write(buffer []byte) (int, error) {
	bytesWritten, err := res.innerWriter.Write(buffer)
	res.contentLength += bytesWritten
	defer func() {
		res.contents.Write(buffer)
	}()
	return bytesWritten, err
}

// Header returns the response headers.
func (res *MockResponseWriter) Header() http.Header {
	return res.headers
}

// WriteHeader sets the status code.
func (res *MockResponseWriter) WriteHeader(statusCode int) {
	res.statusCode = statusCode
}

// InnerResponse returns the backing httpresponse writer.
func (res *MockResponseWriter) InnerResponse() http.ResponseWriter {
	return res
}

// StatusCode returns the status code.
func (res *MockResponseWriter) StatusCode() int {
	return res.statusCode
}

// ContentLength returns the content length.
func (res *MockResponseWriter) ContentLength() int {
	return res.contentLength
}

// Bytes returns the raw response.
func (res *MockResponseWriter) Bytes() []byte {
	return res.contents.Bytes()
}

// Flush is a no-op.
func (res *MockResponseWriter) Flush() {}

// Close is a no-op.
func (res *MockResponseWriter) Close() error {
	return nil
}

// NewMockRequest creates a mock request.
func NewMockRequest(method, path string) *http.Request {
	return &http.Request{
		Method:     method,
		Proto:      "http",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Host:       "localhost:8080",
		RemoteAddr: "127.0.0.1:8080",
		RequestURI: path,
		Header: http.Header{
			HeaderUserAgent: []string{"go-sdk test"},
		},
		URL: &url.URL{
			Scheme:  "http",
			Host:    "localhost",
			Path:    path,
			RawPath: path,
		},
	}
}

// NewMockRequestWithCookie creates a mock request with a cookie attached to it.
func NewMockRequestWithCookie(method, path, cookieName, cookieValue string) *http.Request {
	req := NewMockRequest(method, path)
	req.AddCookie(&http.Cookie{
		Name:   cookieName,
		Domain: "localhost",
		Path:   "/",
		Value:  cookieValue,
	})
	return req
}

// MockSimulateLogin simulates a user login for a given app as mocked request params (i.e. r2 options).
//
// This requires an auth manager to be configured to handle and persist sessions.
func MockSimulateLogin(ctx context.Context, app *App, userID string, opts ...r2.Option) []r2.Option {
	sessionID := NewSessionID()
	session := &Session{
		UserID:     userID,
		SessionID:  sessionID,
		CreatedUTC: time.Now().UTC(),
	}
	if persistHandler, ok := app.AuthPersister.(PersistSessionHandler); ok {
		_ = persistHandler.PersistSession(ctx, session)
	}
	return append([]r2.Option{
		r2.OptCookieValue(app.AuthCookieDefaults.Name, sessionID),
	}, opts...)
}
