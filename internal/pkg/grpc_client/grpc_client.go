package grpc_client

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func New(
	ctx context.Context,
	grpcAddr string,
	backoffLinear time.Duration,
	backoffRetries uint,
	tls *tls.Config,
	grpcOpts ...grpc.DialOption,
) (*grpc.ClientConn, error) {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(backoffLinear)),
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted),
		grpc_retry.WithMax(backoffRetries),
	}

	transportOption := grpc.WithTransportCredentials(insecure.NewCredentials())

	if tls != nil {
		transportOption = grpc.WithTransportCredentials(credentials.NewTLS(tls))
	}

	grpcOpts = append(grpcOpts,
		transportOption,
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)),
	)

	conn, err := grpc.DialContext(ctx, grpcAddr, grpcOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	return conn, nil
}
