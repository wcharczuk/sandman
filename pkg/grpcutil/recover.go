package grpcutil

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type recoveryOptions struct {
	recoveryHandlerFunc RecoveryHandlerFunc
}

// RecoveryOption is a type that provides a recovery option.
type RecoveryOption func(*recoveryOptions)

// OptRecoveryHandler customizes the function for recovering from a panic.
func OptRecoveryHandler(f RecoveryHandlerFunc) RecoveryOption {
	return func(o *recoveryOptions) {
		o.recoveryHandlerFunc = f
	}
}

// RecoveryHandlerFunc is a function that recovers from the panic `p` by returning an `error`.
type RecoveryHandlerFunc func(p any) (err error)

// Recover returns a new unary server interceptor for panic recovery.
func Recover(opts ...RecoveryOption) grpc.UnaryServerInterceptor {
	o := evaluateRecoverOptions(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverFrom(r, o.recoveryHandlerFunc)
			}
		}()
		return handler(ctx, req)
	}
}

func recoverFrom(p interface{}, r RecoveryHandlerFunc) error {
	if r == nil {
		return status.Error(codes.Internal, fmt.Sprint(p))
	}
	return r(p)
}

var (
	defaultRecoverOptions = &recoveryOptions{
		recoveryHandlerFunc: nil,
	}
)

func evaluateRecoverOptions(opts []RecoveryOption) *recoveryOptions {
	optCopy := &recoveryOptions{}
	*optCopy = *defaultRecoverOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}
