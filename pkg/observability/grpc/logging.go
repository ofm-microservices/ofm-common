// Package grpc provides transport-level gRPC diagnostics for OFM services.
package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor logs every unary RPC in development mode. Production
// logs retain method, status, and duration but never serialize request data.
func UnaryServerInterceptor(lg logging.Logger) grpcpkg.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpcpkg.UnaryServerInfo, handler grpcpkg.UnaryHandler) (any, error) {
		started := time.Now()
		resp, err := handler(ctx, req)
		fields := []logging.Field{
			logging.String("rpc_method", info.FullMethod),
			logging.DurationMS(time.Since(started)),
			logging.Int("grpc_code", int(status.Code(err))),
		}
		// Keep payloads out of logs even in dev; transport metadata and types are
		// sufficient for correlation without leaking credentials or PII.
		if logging.IsVerbose(lg) {
			fields = append(fields, logging.String("request_type", fmt.Sprintf("%T", req)), logging.String("response_type", fmt.Sprintf("%T", resp)))
		}
		if err != nil {
			fields = append(fields, logging.Err(err))
			lg.Warn("grpc request completed", fields...)
		} else {
			lg.Info("grpc request completed", fields...)
		}
		return resp, err
	}
}
