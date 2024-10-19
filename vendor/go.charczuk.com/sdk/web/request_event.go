package web

import (
	"strconv"
	"strings"
	"time"

	"go.charczuk.com/sdk/log"
)

// NewRequestEvent returns a new request event.
func NewRequestEvent(ctx Context) RequestEvent {

	// for "raw" results we sometimes will just set the content length header.
	var contentLength int
	headerContentLength := ctx.Response().Header().Get(HeaderContentLength)
	if headerContentLength != "" {
		contentLength, _ = strconv.Atoi(headerContentLength)
	} else {
		contentLength = ctx.Response().ContentLength()
	}

	re := RequestEvent{
		RemoteAddr: GetRemoteAddr(ctx.Request()),
		UserAgent:  ctx.Request().UserAgent(),
		Method:     ctx.Request().Method,
		// we use the `.RequestURI` because we sometimes
		// have to munge the `URL.Path` to make static file servers
		// work correctly and that makes request logs look weird.
		URL:             ctx.Request().RequestURI,
		ContentLength:   contentLength,
		ContentType:     ctx.Response().Header().Get(HeaderContentType),
		ContentEncoding: ctx.Response().Header().Get(HeaderContentEncoding),
		StatusCode:      ctx.Response().StatusCode(),
		Elapsed:         ctx.Elapsed(),
		Locale:          ctx.Locale(),
	}
	if ctx.Route() != nil {
		re.Route = ctx.Route().String()
	}
	return re
}

// RequestEvent is an event type for http requests.
type RequestEvent struct {
	RemoteAddr      string
	UserAgent       string
	Method          string
	URL             string
	Route           string
	ContentLength   int
	ContentType     string
	ContentEncoding string
	StatusCode      int
	Elapsed         time.Duration
	Locale          string
}

// WriteText implements TextWritable.
func (e RequestEvent) String() string {
	wr := new(strings.Builder)

	if e.RemoteAddr != "" {
		wr.WriteString(e.RemoteAddr + " ")
	}
	wr.WriteString(e.Method + " ")
	wr.WriteString(e.URL + " ")
	wr.WriteString(strconv.Itoa(e.StatusCode) + " ")
	wr.WriteString(e.Elapsed.String())
	if e.ContentType != "" {
		wr.WriteString(" " + e.ContentType)
	}
	wr.WriteString(" " + FormatFileSize(e.ContentLength))
	wr.WriteString(" " + e.Locale)
	return wr.String()
}

// FormatForLogger formats the event for use in a logger.
func (e RequestEvent) Attrs() []log.Attr {
	return []log.Attr{
		log.String("remote_addr", e.RemoteAddr),
		log.String("method", e.Method),
		log.String("url", e.URL),
		log.Int("status_code", e.StatusCode),
		log.Duration("elapsed", e.Elapsed),
		log.String("content_type", e.ContentType),
		log.String("content_length", FormatFileSize(e.ContentLength)),
		log.String("locale", e.Locale),
	}
}
