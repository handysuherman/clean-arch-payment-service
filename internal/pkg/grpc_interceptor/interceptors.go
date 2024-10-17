package grpc_interceptor

import (
	"context"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type InterceptorManager interface {
	GrpcLogger(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error)
	ClientRequestLoggerInterceptor() func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error
}

type interceptorManager struct {
	log logger.Logger
}

func New(log logger.Logger) *interceptorManager {
	return &interceptorManager{log: log}
}

func (im *interceptorManager) GrpcLogger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	startTime := time.Now()
	result, err := handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	logger := log.Info()
	if err != nil {
		logger = log.Error().Err(err)
	}

	logger.
		Str(constants.Protocol, constants.GRPC).
		Str(constants.Method, info.FullMethod).
		Int(constants.StatusCode, int(statusCode)).
		Str(constants.StatusText, statusCode.String()).
		Dur(constants.Duration, duration).
		Msg("received a gRPC request")

	return result, err
}

func (im *interceptorManager) ClientRequestLoggerInterceptor() func(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		md, _ := metadata.FromIncomingContext(ctx)
		im.log.GrpcClientInterceptorLogger(method, req, reply, time.Since(startTime), md, err)
		return err
	}
}
