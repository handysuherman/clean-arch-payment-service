package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xendit/xendit-go/v5"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	postgres "github.com/handysuherman/clean-arch-payment-service/internal/pkg/databases/postgresql"
	redisDb "github.com/handysuherman/clean-arch-payment-service/internal/pkg/databases/redis"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
)

var (
	testStore    Repository
	cfg          *config.App
	tlog         logger.Logger
	pqConn       *pgxpool.Pool
	xenditClient *xendit.APIClient
	rConn        redis.UniversalClient
)

func TestMain(m *testing.M) {
	logger := logger.NewLogger()
	cm := config.NewManager(logger, 15*time.Second)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	cfgs, err := cm.Bootstrap(fmt.Sprintf("%v/%s", findModuleRoot(cwd), "etcd-config.yaml"))
	if err != nil {
		logger.Debug(err)
		return
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg = cfgs

	tlog = logger

	psqlOpt := &postgres.Config{
		Host:      cfg.Databases.PostgreSQL.Host,
		Port:      cfg.Databases.PostgreSQL.Port,
		User:      cfg.Databases.PostgreSQL.Username,
		DBName:    cfg.Databases.PostgreSQL.DBName,
		Password:  cfg.Databases.PostgreSQL.Password,
		EnableTls: cfg.Databases.PostgreSQL.EnableTLS,
	}

	if cfg.Databases.PostgreSQL.EnableTLS {
		pqsqlTls, err := helper.Base64EncodedTLS(cfg.TLS.PostgreSQL.Ca, cfg.TLS.PostgreSQL.Cert, cfg.TLS.PostgreSQL.Key)
		if err != nil {
			tlog.Info(err)
		}

		pqsqlTls.ServerName = cfg.Databases.PostgreSQL.Host

		psqlOpt.TLs = pqsqlTls
	}

	ctx := context.Background()

	pgxConn, err := postgres.NewPgxConn(ctx, psqlOpt)
	if err != nil {
		tlog.Error(err)
		return
	}
	pqConn = pgxConn
	defer pgxConn.Close()

	redisOpt := redisDb.Config{
		Host:      cfg.Databases.Redis.Servers[0],
		Password:  cfg.Databases.Redis.Password,
		DB:        cfg.Databases.Redis.DB,
		PoolSize:  cfg.Databases.Redis.PoolSize,
		TLsEnable: cfg.Databases.Redis.EnableTLS,
	}

	if cfg.Databases.Redis.EnableTLS {
		redisTls, err := helper.Base64EncodedTLS(cfg.TLS.Redis.Ca, cfg.TLS.Redis.Cert, cfg.TLS.Redis.Key)
		if err != nil {
			tlog.Error(err)
			return
		}

		redisOpt.TLs = redisTls
	}

	redisConn, _ := redisDb.NewUniversalRedisClient(ctx, &redisOpt)
	rConn = redisConn

	xenditClient = xendit.NewClient(cfg.Services.Internal.PaymentGatewayKeys.Development)

	testStore = NewStore(logger, cfg, pgxConn, redisConn, xenditClient)
	os.Exit(m.Run())
}
