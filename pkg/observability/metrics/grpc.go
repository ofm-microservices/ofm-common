package metrics

import (
	"context"
	"path"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor records low-cardinality metrics for gRPC server calls.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		started := time.Now()
		resp, err := handler(ctx, req)
		service, method := splitFullMethod(info.FullMethod)
		Global().ObserveGRPCServer(service, method, status.Code(err).String(), time.Since(started))
		return resp, err
	}
}

// UnaryClientInterceptor records low-cardinality metrics for outbound gRPC calls.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		started := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		service, name := splitFullMethod(method)
		Global().ObserveGRPCClient(service, name, status.Code(err).String(), time.Since(started))
		return err
	}
}

func splitFullMethod(fullMethod string) (string, string) {
	service := strings.TrimPrefix(path.Dir(fullMethod), "/")
	method := path.Base(fullMethod)
	if service == "." {
		service = "unknown"
	}
	return service, method
}
