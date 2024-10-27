package viewmodel

import (
	"encoding/base64"
	"time"

	v1 "sandman/proto/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Timer struct {
	Name     string            `yaml:"name"`
	Labels   map[string]string `yaml:"labels,omitempty"`
	Priority uint32            `yaml:"priority"`
	DueUTC   time.Time         `yaml:"due_utc"`
	Hook     Hook              `yaml:"hook"`
}

func (t Timer) ToProto() *v1.Timer {
	bodyData, _ := base64.StdEncoding.DecodeString(t.Hook.Body)
	return &v1.Timer{
		Name:        t.Name,
		Labels:      t.Labels,
		Priority:    t.Priority,
		DueUtc:      timestamppb.New(t.DueUTC),
		HookUrl:     t.Hook.URL,
		HookMethod:  t.Hook.Method,
		HookHeaders: t.Hook.Headers,
		HookBody:    bodyData,
	}
}

type Hook struct {
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
}
