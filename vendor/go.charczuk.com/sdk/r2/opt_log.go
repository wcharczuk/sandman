package r2

import (
	"net/http"
	"time"

	"go.charczuk.com/sdk/log"
)

// OptLog configures the request to use logging.
func OptLog(logger *log.Logger) Option {
	return func(r *Request) error {
		r.onResponse = append(r.onResponse, func(req *http.Request, res *http.Response, start time.Time, err error) error {
			logger.WithGroup("r2_request").Info("request",
				log.String("method", req.Method),
				log.String("url", req.URL.String()),
				log.Int("status_code", res.StatusCode),
				log.Duration("elapsed", time.Now().UTC().Sub(start)),
			)
			return nil
		})
		return nil
	}
}
