package r2

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.charczuk.com/sdk/errutil"
)

// New returns a new request.
//
// The default method is GET, the default Proto is `HTTP/1.1`.
func New(remoteURL string, options ...Option) *Request {
	var r Request
	u, err := url.Parse(remoteURL)
	if err != nil {
		r.err = err
		return &r
	}
	u.Host = RemoveHostEmptyPort(u.Host)
	r.req = &http.Request{
		Method:     http.MethodGet,
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       u.Host,
	}
	for _, option := range options {
		if err = option(&r); err != nil {
			r.err = err
			return &r
		}
	}
	return &r
}

// Request is a combination of the http.Request options and the underlying client.
type Request struct {
	req        *http.Request
	err        error
	client     *http.Client
	closer     func() error
	onRequest  []func(*http.Request) error
	onResponse []func(*http.Request, *http.Response, time.Time, error) error
}

// WithContext implements the `WithContext` method for the underlying request.
//
// It is preserved here because the pointer indirects are non-trivial.
func (r *Request) WithContext(ctx context.Context) *Request {
	*r.req = *r.req.WithContext(ctx)
	return r
}

// Method returns the request method.
func (r *Request) Method() string {
	return r.req.Method
}

// URL returns the request url.
func (r *Request) URL() *url.URL {
	return r.req.URL
}

// Do executes the request.
func (r Request) Do() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}
	if len(r.req.PostForm) > 0 {
		if r.req.Body != nil {
			return nil, ErrFormAndBodySet
		}
		body := r.req.PostForm.Encode()
		buffer := bytes.NewBufferString(body)
		r.req.ContentLength = int64(buffer.Len())
		r.req.Body = io.NopCloser(buffer)
	}

	if r.req.Body == nil {
		r.req.Body = http.NoBody
		r.req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(http.NoBody), nil }
	}

	started := time.Now().UTC()
	for _, listener := range r.onRequest {
		if err := listener(r.req); err != nil {
			return nil, err
		}
	}

	var err error
	var res *http.Response
	if r.client != nil {
		res, err = r.client.Do(r.req)
	} else {
		res, err = http.DefaultClient.Do(r.req)
	}
	for _, listener := range r.onResponse {
		if listenerErr := listener(r.req, res, started, err); listenerErr != nil {
			return nil, listenerErr
		}
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Close closes the request if there is a closer specified.
func (r *Request) Close() error {
	if r.closer != nil {
		return r.closer()
	}
	return nil
}

// Discard reads the response fully and discards all data it reads, and returns the response metadata.
func (r Request) Discard() (res *http.Response, err error) {
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = errutil.Append(err, closeErr)
		}
	}()
	res, err = r.Do()
	if err != nil {
		res = nil
		return
	}
	defer res.Body.Close()
	_, err = io.Copy(io.Discard, res.Body)
	return
}

// CopyTo copies the response body to a given writer.
func (r Request) CopyTo(dst io.Writer) (count int64, err error) {
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = errutil.Append(err, closeErr)
		}
	}()

	var res *http.Response
	res, err = r.Do()
	if err != nil {
		res = nil
		return
	}
	defer res.Body.Close()
	count, err = io.Copy(dst, res.Body)
	return
}

// Bytes reads the response and returns it as a byte array, along with the response metadata..
func (r Request) Bytes() (contents []byte, res *http.Response, err error) {
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = errutil.Append(err, closeErr)
		}
	}()
	res, err = r.Do()
	if err != nil {
		res = nil
		return
	}
	defer res.Body.Close()
	contents, err = io.ReadAll(res.Body)
	return
}

// JSON reads the response as json into a given object and returns the response metadata.
func (r Request) JSON(dst interface{}) (res *http.Response, err error) {
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = errutil.Append(err, closeErr)
		}
	}()

	res, err = r.Do()
	if err != nil {
		res = nil
		return
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNoContent {
		err = ErrNoContentJSON
		return
	}
	if err = json.NewDecoder(res.Body).Decode(dst); err != nil {
		return
	}
	return
}

// Gob reads the response as gob into a given object and returns the response metadata.
func (r Request) Gob(dst interface{}) (res *http.Response, err error) {
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = errutil.Append(err, closeErr)
		}
	}()

	res, err = r.Do()
	if err != nil {
		res = nil
		return
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNoContent {
		err = ErrNoContentGob
		return
	}
	if err = gob.NewDecoder(res.Body).Decode(dst); err != nil {
		return
	}
	return
}
