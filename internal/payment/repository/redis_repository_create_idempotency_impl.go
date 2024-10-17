package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/handysuherman/clean-arch-payment-service/internal/pb"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/unierror/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

var (
	redisCreatePaymentIdempotencyPrefixKey = "reader:create_payment"
)

func (r *RedisRepositoryImpl) PutCreatePaymentIdempotencyKey(ctx context.Context, key string, arg *pb.CreatePaymentResponse) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RedisRepository.PutCreatePaymentIdempotencyKey")
	defer span.Finish()

	payload, err := proto.Marshal(arg)
	if err != nil {
		r.log.Warnf("serializer.marshal.err: %v", err)
		return
	}

	prefixKey := helper.RedisPrefixes(
		key,
		redisCreatePaymentIdempotencyPrefixKey,
		r.cfg.Databases.Redis.Prefixes.CreatePayment.Prefix,
		r.cfg.Databases.Redis.AppID,
	)

	if err := r.redisClient.Set(
		ctx,
		prefixKey,
		payload,
		r.cfg.Databases.Redis.Prefixes.CreatePayment.ExpirationDuration,
	).Err(); err != nil {
		return
	}

	r.log.Debugf("put-prefix: %s, key: %s", prefixKey, key)
}

func (r *RedisRepositoryImpl) GetCreatePaymentIdempotencyKey(ctx context.Context, key string) (*pb.CreatePaymentResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RedisRepositorytImpl.GetCreatePaymentIdempotencyKey")
	defer span.Finish()

	prefixKey := helper.RedisPrefixes(
		key,
		redisCreatePaymentIdempotencyPrefixKey,
		r.cfg.Databases.Redis.Prefixes.CreatePayment.Prefix,
		r.cfg.Databases.Redis.AppID,
	)

	msg, err := r.redisClient.Get(ctx, prefixKey).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			r.log.Warnf("redis.client.get.err: %v", err)
		}
		return nil, fmt.Errorf("unable to get cache: %w", tracing.TraceWithError(span, err))
	}

	var payload pb.CreatePaymentResponse
	if err := proto.Unmarshal(msg, &payload); err != nil {
		return nil, fmt.Errorf("serializer.Unmarshal.err: %v", tracing.TraceWithError(span, err))
	}

	r.log.Debugf("get-key: %s, key: %s", prefixKey, key)

	return &payload, nil
}

func (r *RedisRepositoryImpl) DeleteCreatePaymentIdempotencyKey(ctx context.Context, key string) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "RedisRepositoryImpl.DeleteCreatePaymentIdempotencyKey")
	defer span.Finish()

	prefixKey := helper.RedisPrefixes(
		key,
		redisCreatePaymentIdempotencyPrefixKey,
		r.cfg.Databases.Redis.Prefixes.CreatePayment.Prefix,
		r.cfg.Databases.Redis.AppID,
	)

	if err := r.redisClient.Del(ctx, prefixKey).Err(); err != nil {
		r.log.Warnf("delete.cache.del.err: %v", err)
		return
	}
	r.log.Debugf("del-prefix: %s, key: %s", prefixKey, key)
}
