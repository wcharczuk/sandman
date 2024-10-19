package db

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/stringutil"
)

// QueryEvent represents a database query.
type QueryEvent struct {
	Database     string        `json:"database,omitempty"`
	Engine       string        `json:"engine,omitempty"`
	Username     string        `json:"username,omitempty"`
	Label        string        `json:"label,omitempty"`
	Body         string        `json:"body,omitempty"`
	Elapsed      time.Duration `json:"elapsed,omitempty"`
	Err          error         `json:"err,omitempty"`
	RowsAffected int64         `json:"rowsAffected,omitempty"`
}

// WriteText writes the event text to the output.
func (e QueryEvent) String() string {
	wr := new(strings.Builder)
	wr.WriteString("[")
	if len(e.Engine) > 0 {
		wr.WriteString(e.Engine + " ")
	}
	if len(e.Username) > 0 {
		wr.WriteString(e.Username + "@")
	}
	wr.WriteString(e.Database + "]")

	if len(e.Label) > 0 {
		wr.WriteString(" [" + e.Label + "]")
	}
	if len(e.Body) > 0 {
		wr.WriteString(" " + stringutil.CompressSpace(e.Body))
	}
	wr.WriteString(" " + e.Elapsed.String())
	if e.RowsAffected > 0 {
		wr.WriteString(" " + strconv.FormatInt(e.RowsAffected, 10) + " rows")
	}
	if e.Err != nil {
		wr.WriteString(" failed")
	}
	return wr.String()
}

// Attrs returns log attributes for the event.
func (e QueryEvent) Attrs() (attrs []log.Attr) {
	if len(e.Label) > 0 {
		attrs = append(attrs, log.String("label", e.Label))
	}
	if len(e.Body) > 0 {
		attrs = append(attrs, log.String("body", `"`+stringutil.CompressSpace(e.Body)+`"`))
	}
	if len(e.Engine) > 0 {
		attrs = append(attrs, log.String("engine", e.Engine))
	}
	if len(e.Username) > 0 {
		attrs = append(attrs, log.String("username", e.Username))
	}
	if len(e.Database) > 0 {
		attrs = append(attrs, log.String("database", e.Database))
	}
	if e.Elapsed > 0 {
		attrs = append(attrs, log.String("elapsed", e.Elapsed.String()))
	}
	if e.RowsAffected > 0 {
		attrs = append(attrs, log.String("rows_affected", fmt.Sprint(e.RowsAffected)))
	}
	if e.Err != nil {
		attrs = append(attrs, log.Bool("failed", true))
	}
	return
}
