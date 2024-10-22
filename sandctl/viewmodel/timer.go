package viewmodel

import (
	"encoding/base64"
	"time"

	v1 "sandman/proto/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Timer struct {
	Name   string            `yaml:"name"`
	Labels map[string]string `yaml:"labels,omitempty"`
	DueUTC time.Time         `yaml:"due_utc"`
	RPC    RPC               `yaml:"rpc"`
}

func (t Timer) ToProto() *v1.Timer {
	argsData, _ := base64.StdEncoding.DecodeString(t.RPC.ArgsData)
	return &v1.Timer{
		Name:             t.Name,
		Labels:           t.Labels,
		DueUtc:           timestamppb.New(t.DueUTC),
		RpcAddr:          t.RPC.Addr,
		RpcAuthority:     t.RPC.Authority,
		RpcMethod:        t.RPC.Method,
		RpcMeta:          t.RPC.Meta,
		RpcArgsTypeUrl:   t.RPC.ArgsTypeURL,
		RpcArgsData:      argsData,
		RpcReturnTypeUrl: t.RPC.ReturnTypeURL,
	}
}

type RPC struct {
	Addr          string            `yaml:"addr"`
	Authority     string            `yaml:"authority,omitempty"`
	Method        string            `yaml:"method"`
	Meta          map[string]string `yaml:"meta,omitempty"`
	ArgsTypeURL   string            `yaml:"argsTypeURL,omitempty"`
	ArgsData      string            `yaml:"argsData,omitempty"`
	ReturnTypeURL string            `yaml:"returnTypeURL,omitempty"`
}
