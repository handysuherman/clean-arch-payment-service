package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/serializer"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
)

const (
	redisPrefixKey = "reader:payment"
)

func (r *RedisRepositoryImpl) PutCache(ctx context.Context, arg *PaymentMethod) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RedisRepositoryImpl.PutCache")
	defer span.Finish()

	key := fmt.Sprintf("%s-%s", arg.PaymentCustomerID, arg.PaymentMethodID)

	payload, err := serializer.Marshal(arg)
	if err != nil {
		r.log.Warnf("serializer.marshal.err: %v", err)
		return
	}

	prefixKey := helper.RedisPrefixes(
		key,
		redisPrefixKey,
		r.cfg.Databases.Redis.Prefixes.Payment.Prefix,
		r.cfg.Databases.Redis.AppID,
	)

	if err := r.redisClient.Set(ctx, prefixKey, payload, r.cfg.Databases.Redis.Prefixes.Payment.ExpirationDuration).Err(); err != nil {
		return
	}

	r.log.Debugf("put-prefix: %s, key: %s", prefixKey, key)
}

func (r *RedisRepositoryImpl) GetCache(ctx context.Context, paymentCustomerID string, paymentMethodID string) (*PaymentMethod, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RedisRepositorytImpl.GetCache")
	defer span.Finish()

	key := fmt.Sprintf("%s-%s", paymentCustomerID, paymentMethodID)

	prefixKey := helper.RedisPrefixes(
		key,
		redisPrefixKey,
		r.cfg.Databases.Redis.Prefixes.Payment.Prefix,
		r.cfg.Databases.Redis.AppID,
	)

	msg, err := r.redisClient.Get(ctx, prefixKey).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			r.log.Warnf("redis.client.get.err: %v", err)
		}
		return nil, fmt.Errorf("unable to get cache: %w", tracing.TraceWithError(span, err))
	}

	var payload PaymentMethod
	if err := serializer.Unmarshal(msg, &payload); err != nil {
		return nil, fmt.Errorf("serializer.Unmarshal.err: %v", tracing.TraceWithError(span, err))
	}

	r.log.Debugf("get-prefix: %s, key: %s", prefixKey, key)

	return &payload, nil
}

func (r *RedisRepositoryImpl) DeleteCache(ctx context.Context, paymentCustomerID string, paymentMethodID string) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RedisRepositoryImpl.DeleteCache")
	defer span.Finish()

	key := fmt.Sprintf("%s-%s", paymentCustomerID, paymentMethodID)

	prefixKey := helper.RedisPrefixes(
		key,
		redisPrefixKey,
		r.cfg.Databases.Redis.Prefixes.Payment.Prefix,
		r.cfg.Databases.Redis.AppID,
	)

	if err := r.redisClient.Del(ctx, prefixKey).Err(); err != nil {
		r.log.Warnf("delete.cache.del.err: %v", err)
		return
	}
	r.log.Debugf("del-prefix: %s, key: %s", prefixKey, key)
}
