package r2

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
	"net/http"
)

// OptBody sets a body on a context from bytes.
func OptBody(body []byte) Option {
	return func(r *Request) error {
		r.req.ContentLength = int64(len(body))
		r.req.Body = io.NopCloser(bytes.NewReader(body))
		r.req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(body)), nil
		}
		r.req.ContentLength = int64(len(body))
		return nil
	}
}

// OptBodyReader sets the post body on the request.
func OptBodyReader(contents io.ReadCloser) Option {
	return func(r *Request) error {
		r.req.Body = contents
		return nil
	}
}

// OptBodyJSON sets the post body on the request to a given
// value encoded with json.
func OptBodyJSON(obj interface{}) Option {
	return func(r *Request) error {
		contents, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		r.req.Body = io.NopCloser(bytes.NewReader(contents))
		r.req.GetBody = func() (io.ReadCloser, error) {
			r := bytes.NewReader(contents)
			return io.NopCloser(r), nil
		}
		r.req.ContentLength = int64(len(contents))
		if r.req.Header == nil {
			r.req.Header = make(http.Header)
		}
		r.req.Header.Set("Content-Type", "application/json; charset=utf-8")
		return nil
	}
}

// OptBodyGob sets the post body on the request to a given
// value encoded with gob.
func OptBodyGob(obj interface{}) Option {
	return func(r *Request) error {
		contents := new(bytes.Buffer)
		err := gob.NewEncoder(contents).Encode(obj)
		if err != nil {
			return err
		}
		r.req.Body = io.NopCloser(contents)
		r.req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(contents), nil
		}
		r.req.ContentLength = int64(contents.Len())
		if r.req.Header == nil {
			r.req.Header = make(http.Header)
		}
		r.req.Header.Set("Content-Type", "application/gob")
		return nil
	}
}
