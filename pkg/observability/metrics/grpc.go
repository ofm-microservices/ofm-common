package metrics

import (
	"context"
	"path"
	"strings"
	"time"

	requestmetadata "github.com/ofm-microservices/ofm-common/pkg/observability/metadata"
	"github.com/ofm-microservices/ofm-common/pkg/resilience"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor records low-cardinality metrics for gRPC server calls.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return requestmetadata.UnaryServerInterceptor()(ctx, req, info, func(metadataCtx context.Context, metadataReq any) (any, error) {
			started := time.Now()
			resp, err := handler(metadataCtx, metadataReq)
			service, method := splitFullMethod(info.FullMethod)
			Global().ObserveGRPCServer(service, method, status.Code(err).String(), time.Since(started))
			return resp, err
		})
	}
}

// UnaryClientInterceptor records low-cardinality metrics for outbound gRPC calls.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return ResilientUnaryClientInterceptor(resilience.BreakerConfigFromEnv())
}

// ResilientUnaryClientInterceptor records metrics and protects one outbound
// gRPC connection with a circuit breaker and per-call timeout.
func ResilientUnaryClientInterceptor(config resilience.BreakerConfig) grpc.UnaryClientInterceptor {
	breaker := resilience.NewBreaker(config)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		started := time.Now()
		err := breaker.DoClassified(ctx, func(callCtx context.Context) error {
			return requestmetadata.UnaryClientInterceptor()(callCtx, method, req, reply, cc, invoker, opts...)
		}, resilience.IsTransientDependencyError)
		service, name := splitFullMethod(method)
		Global().ObserveGRPCClient(service, name, status.Code(err).String(), time.Since(started))
		Global().ObserveCircuitBreaker(service, string(breaker.State()))
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
