package web

import (
	"net/http"

	"go.charczuk.com/sdk/errutil"
)

// Raw returns a new raw result.
func Raw(contents []byte) *RawResult {
	return &RawResult{
		StatusCode:  http.StatusOK,
		ContentType: http.DetectContentType(contents),
		Response:    contents,
	}
}

// RawString returns a new text result.
func RawString(contents string) *RawResult {
	return &RawResult{
		StatusCode:  http.StatusOK,
		ContentType: ContentTypeText,
		Response:    []byte(contents),
	}
}

// RawWithContentType returns a binary response with a given content type.
func RawWithContentType(contentType string, body []byte) *RawResult {
	return &RawResult{
		StatusCode:  http.StatusOK,
		ContentType: contentType,
		Response:    body,
	}
}

// RawWithErr returns a binary response with a given application error.
func RawWithErr(contentType string, body []byte, err error) *RawResult {
	return &RawResult{
		StatusCode:  http.StatusOK,
		ContentType: contentType,
		Response:    body,
		Err:         err,
	}
}

// RawResult is for when you just want to dump bytes.
type RawResult struct {
	StatusCode  int
	ContentType string
	Response    []byte
	Err         error
}

// Render renders the result.
func (rr *RawResult) Render(ctx Context) error {
	if rr.ContentType != "" {
		ctx.Response().Header().Set(HeaderContentType, rr.ContentType)
	}
	ctx.Response().WriteHeader(rr.StatusCode)
	_, err := ctx.Response().Write(rr.Response)
	return errutil.Append(err, rr.Err)
}
