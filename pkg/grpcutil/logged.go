package grpcutil

import (
	"context"
	"time"

	"go.charczuk.com/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Logged returns a unary server interceptor.
func Logged(logger *log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, args interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now().UTC()
		result, err := handler(ctx, args)
		if logger != nil {
			attrs := []any{
				log.String("method", info.FullMethod),
				log.String("elapsed", time.Since(startTime).Round(time.Microsecond).String()),
			}
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				attrs = append(attrs,
					log.String("authority", rpcMetaValue(md, MetaKeyAuthority)),
				)
			}
			logger.Info("rpc", attrs...)
		}
		return result, err
	}
}

// MetaKeys
const (
	MetaKeyAuthority   = "authority"
	MetaKeyUserAgent   = "user-agent"
	MetaKeyContentType = "content-type"
)

func rpcMetaValue(md metadata.MD, key string) string {
	if values, ok := md[key]; ok {
		if len(values) > 0 {
			return values[0]
		}
	}
	return ""
}
