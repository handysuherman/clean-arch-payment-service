package repository

import (
	"context"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type RedisRepository interface {
	PutCreatePaymentIdempotencyKey(ctx context.Context, key string, arg *pb.CreatePaymentResponse)
	GetCreatePaymentIdempotencyKey(ctx context.Context, key string) (*pb.CreatePaymentResponse, error)
	DeleteCreatePaymentIdempotencyKey(ctx context.Context, key string)

	PutCache(ctx context.Context, arg *PaymentMethod)
	GetCache(ctx context.Context, paymentCustomerID string, paymentMethodID string) (*PaymentMethod, error)
	DeleteCache(ctx context.Context, paymentCustomerID string, paymentMethodID string)

	PutCustomerCache(ctx context.Context, arg *Customer)
	GetCustomerCache(ctx context.Context, key string) (*Customer, error)
	DeleteCustomerCache(ctx context.Context, key string)
}

type RedisRepositoryImpl struct {
	log         logger.Logger
	cfg         *config.App
	redisClient redis.UniversalClient
}

var _ RedisRepository = (*RedisRepositoryImpl)(nil)

func NewRedisRepositoryImpl(log logger.Logger, cfg *config.App, redisClient redis.UniversalClient) *RedisRepositoryImpl {
	return &RedisRepositoryImpl{
		log:         log,
		cfg:         cfg,
		redisClient: redisClient,
	}
}
