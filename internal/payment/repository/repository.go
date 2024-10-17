package repository

import (
	"context"
	"fmt"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/xendit/xendit-go/v5"
)

type Repository interface {
	Querier
	PaymentProvider
	RedisRepository

	CreateCustomerTx(ctx context.Context, arg *CreateCustomerTxParams) (CreateCustomerTxResult, error)
	CreateEwalletPaymentTx(ctx context.Context, arg *CreateEwalletPaymentTxParams) (CreateEwalletPaymentTxResult, error)
	CreateQrCodePaymentTx(ctx context.Context, arg *CreateQrCodePaymentTxParams) (CreateQrCodePaymentTxResult, error)
	CreateVirtualAccountBankPaymentTx(ctx context.Context, arg *CreateVirtualAccountBankPaymentTxParams) (CreateVirtualAccountBankPaymentTxResult, error)
	UpdateTx(ctx context.Context, arg *UpdateTxParams) (UpdateTxResult, error)

	OnConfigUpdate(key string, config *config.App)
	OnPqsqlUpdate(key string, pqsqlConnection *pgxpool.Pool)
	OnRedisUpdate(key string, redisConnection redis.UniversalClient)
}

type Store struct {
	log logger.Logger
	cfg *config.App
	db  *pgxpool.Pool
	*Queries
	*PaymentProviderImpl
	*RedisRepositoryImpl
}

func NewStore(
	log logger.Logger,
	cfg *config.App,
	db *pgxpool.Pool,
	redisClient redis.UniversalClient,
	xenditClient *xendit.APIClient,
) Repository {
	log = log.WithPrefix(fmt.Sprintf("%s-%s", "payment", constants.Repository))
	return &Store{
		log:                 log,
		cfg:                 cfg,
		db:                  db,
		Queries:             New(db),
		PaymentProviderImpl: NewPaymentProviderImpl(log, cfg, xenditClient),
		RedisRepositoryImpl: NewRedisRepositoryImpl(log, cfg, redisClient),
	}
}
