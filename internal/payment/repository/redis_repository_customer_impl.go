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
	redisCustomerPrefixKey = "reader:payment_customer"
)

func (r *RedisRepositoryImpl) PutCustomerCache(ctx context.Context, arg *Customer) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RedisRepositoryImpl.PutCustomerCache")
	defer span.Finish()

	payload, err := serializer.Marshal(arg)
	if err != nil {
		r.log.Warnf("serializer.marshal.err: %v", err)
		return
	}

	prefixKey := helper.RedisPrefixes(
		arg.CustomerAppID,
		redisCustomerPrefixKey,
		r.cfg.Databases.Redis.Prefixes.Customer.Prefix,
		r.cfg.Databases.Redis.AppID,
	)

	if err := r.redisClient.Set(ctx, prefixKey, payload, r.cfg.Databases.Redis.Prefixes.Customer.ExpirationDuration).Err(); err != nil {
		return
	}

	r.log.Debugf("put-prefix: %s, key: %s", prefixKey, arg.CustomerAppID)
}

func (r *RedisRepositoryImpl) GetCustomerCache(ctx context.Context, key string) (*Customer, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RedisRepositorytImpl.GetCustomerCache")
	defer span.Finish()

	prefixKey := helper.RedisPrefixes(key, redisCustomerPrefixKey, r.cfg.Databases.Redis.Prefixes.Customer.Prefix, r.cfg.Databases.Redis.AppID)

	msg, err := r.redisClient.Get(ctx, prefixKey).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			r.log.Warnf("redis.client.get.err: %v", err)
		}
		return nil, fmt.Errorf("unable to get cache: %w", tracing.TraceWithError(span, err))
	}

	var payload Customer
	if err := serializer.Unmarshal(msg, &payload); err != nil {
		return nil, fmt.Errorf("serializer.Unmarshal.err: %v", tracing.TraceWithError(span, err))
	}

	r.log.Debugf("get-key: %s, key: %s", prefixKey, key)

	return &payload, nil
}

func (r *RedisRepositoryImpl) DeleteCustomerCache(ctx context.Context, key string) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RedisRepositoryImpl.DeleteCustomerCache")
	defer span.Finish()

	prefixKey := helper.RedisPrefixes(
		key,
		redisCustomerPrefixKey,
		r.cfg.Databases.Redis.Prefixes.Customer.Prefix,
		r.cfg.Databases.Redis.AppID,
	)

	if err := r.redisClient.Del(ctx, prefixKey).Err(); err != nil {
		r.log.Warnf("delete.cache.del.err: %v", err)
		return
	}
	r.log.Debugf("del-prefix: %s, key: %s", prefixKey, key)
}
