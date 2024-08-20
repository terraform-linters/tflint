package interceptor

import (
	"context"

	"github.com/terraform-linters/tflint-plugin-sdk/logger"
	"google.golang.org/grpc"
)

// RequestLogging is an interceptor for gRPC request logging.
// It outouts all request logs as "trace" level, and if an error occurs,
// it outputs the response as "error" level.
func RequestLogging(direction string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Trace("gRPC request", "direction", direction, "method", info.FullMethod, "req", req)
		ret, err := handler(ctx, req)
		if err != nil {
			logger.Error("failed to gRPC request", "direction", direction, "method", info.FullMethod, "err", err)
		}
		return ret, err
	}
}
