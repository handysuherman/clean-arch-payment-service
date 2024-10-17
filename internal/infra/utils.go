package infra

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	waitShutdownDur = 3 * time.Second
)

func (a *app) loadTLSCredentials() (credentials.TransportCredentials, error) {
	ca, err := base64.StdEncoding.DecodeString(a.cfg.TLS.App.Ca)
	if err != nil {
		a.log.Errorf("error while decoding TLSCredentials ca: %v", err)
		return nil, err
	}

	cert, err := base64.StdEncoding.DecodeString(a.cfg.TLS.App.Cert)
	if err != nil {
		a.log.Errorf("error while decoding TLSCredentials cert: %v", err)
		return nil, err
	}

	key, err := base64.StdEncoding.DecodeString(a.cfg.TLS.App.Key)
	if err != nil {
		a.log.Errorf("error while ecoding TLSCredentials key: %v", err)
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	certs, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load key part: %v", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{certs},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}

func (a *app) loadConsulCerts() (*consul.ConsulClientCerts, error) {
	ca, err := base64.StdEncoding.DecodeString(a.cfg.TLS.Consul.Ca)
	if err != nil {
		a.log.Errorf("error while decoding consulConn ca: %v", err)
		return nil, err
	}

	cert, err := base64.StdEncoding.DecodeString(a.cfg.TLS.Consul.Cert)
	if err != nil {
		a.log.Errorf("error while decoding consulConn cert: %v", err)
		return nil, err
	}

	key, err := base64.StdEncoding.DecodeString(a.cfg.TLS.Consul.Key)
	if err != nil {
		a.log.Errorf("error while ecoding consulConn key: %v", err)
		return nil, err
	}

	return &consul.ConsulClientCerts{
		Ca:   ca,
		Cert: cert,
		Key:  key,
	}, nil
}

func (a *app) waitGracefulShutdown(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		a.doneCh <- struct{}{}
	}()
}

func (a *app) configurationPriorities() config.KeyPriorities {
	kp := make(config.KeyPriorities)

	kp[a.cfg.Etcd.Keys.Configurations.Postgresql] = 0
	kp[a.cfg.Etcd.Keys.Configurations.Redis] = 0
	kp[a.cfg.Etcd.Keys.Configurations.Brokers] = 0
	kp[a.cfg.Etcd.Keys.TLS.Postgresql] = 0
	kp[a.cfg.Etcd.Keys.TLS.Redis] = 0
	kp[a.cfg.Etcd.Keys.TLS.Kafka] = 0
	kp[a.cfg.Etcd.Keys.TLS.Paseto] = 0
	kp[a.cfg.Etcd.Keys.TLS.Consul] = 0

	return kp
}

func (a *app) shutdownProcess(ctx context.Context, grpcServer *grpc.Server) {
	<-ctx.Done()
	a.waitGracefulShutdown(waitShutdownDur)

	if err := a.metricsServer.Shutdown(ctx); err != nil {
		a.log.Infof("a.metricsServer.Shutdown.err: %v", err)
	}

	if err := a.gracefulShutDownHealthCheckServer(ctx); err != nil {
		a.log.Infof("a.gracefulShutdownHealthCheckServer.err: %v", err)
	}

	<-a.doneCh

	grpcServer.GracefulStop()
	a.log.Info("server shutdowned successfully..,")
}
