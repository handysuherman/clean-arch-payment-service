package infra

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grpc2 "github.com/handysuherman/clean-arch-payment-service/internal/payment/delivery/grpc"
	healthGrpc "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	maxConnectionIdle = 5
	gRPCTimeout       = 30
	maxConnectionAge  = 5
	gRPCTime          = 10
)

func (a *app) newGrpcServer(ctx context.Context) (func() error, *grpc.Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", a.cfg.Services.Internal.Port))
	if err != nil {
		return nil, nil, errors.Wrap(err, "net.Listen")
	}

	serverOptions := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: maxConnectionIdle * time.Minute,
			MaxConnectionAge:  maxConnectionAge * time.Minute,
			Timeout:           gRPCTimeout * time.Second,
			Time:              gRPCTime * time.Minute,
		}),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_recovery.UnaryServerInterceptor(),
			a.im.GrpcLogger,
		)),
	}

	if a.cfg.Services.Internal.EnableTLS {
		tlsCredentials, err := a.loadTLSCredentials()
		if err != nil {
			return l.Close, nil, fmt.Errorf("cannot load TLS Credentials: %v", err)
		}
		serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))
	}

	hs := health.NewServer()

	go func() {
		defer hs.SetServingStatus("", healthGrpc.HealthCheckResponse_NOT_SERVING)
		defer hs.SetServingStatus(pb.PaymentService_ServiceDesc.ServiceName, healthGrpc.HealthCheckResponse_NOT_SERVING)

		a.log.Info("Health check goroutine started.")
		defer a.log.Info("Health check goroutine stopped.")

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.toggleHealthStatus(hs)
			}
		}
	}()

	grpcServer := grpc.NewServer(serverOptions...)
	deliveryGrpc := grpc2.New(
		a.log,
		a.cfg,
		a.v,
		a.usecase,
		a.metrics,
	)
	a.cfgManager.RegisterObserver(deliveryGrpc, 3)
	pb.RegisterPaymentServiceServer(grpcServer, deliveryGrpc)
	grpc_prometheus.Register(grpcServer)

	if a.cfg.Services.Internal.Environment != "production" {
		reflection.Register(grpcServer)
	}

	go func() {
		a.log.Infof("gRPC server is listening on port: %v, TLS = %v", a.cfg.Services.Internal.Port, a.cfg.Services.Internal.EnableTLS)
		a.log.Fatal(grpcServer.Serve(l))
	}()

	return l.Close, grpcServer, nil
}

func (a *app) toggleHealthStatus(hs *health.Server) {
	status := healthGrpc.HealthCheckResponse_SERVING
	hs.SetServingStatus("", status)
	hs.SetServingStatus(pb.PaymentService_ServiceDesc.ServiceName, status)
}
