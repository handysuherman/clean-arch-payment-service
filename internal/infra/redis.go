package infra

import (
	"context"

	redisDb "github.com/handysuherman/clean-arch-payment-service/internal/pkg/databases/redis"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
)

func (a *app) redis(ctx context.Context) error {
	redisOpt := redisDb.Config{
		Host:      a.cfg.Databases.Redis.Servers[0],
		Password:  a.cfg.Databases.Redis.Password,
		DB:        a.cfg.Databases.Redis.DB,
		PoolSize:  a.cfg.Databases.Redis.PoolSize,
		TLsEnable: a.cfg.Databases.Redis.EnableTLS,
	}

	if redisOpt.TLsEnable {
		redisTls, err := helper.Base64EncodedTLS(a.cfg.TLS.Redis.Ca, a.cfg.TLS.Redis.Cert, a.cfg.TLS.Redis.Key)
		if err != nil {
			a.log.Warnf("redis.helper.base64encodedtls.err: %v", err)
			return err
		}

		redisOpt.TLs = redisTls
	}

	redisConn, err := redisDb.NewUniversalRedisClient(ctx, &redisOpt)
	if err != nil {
		a.log.Warnf("redis.redisdb.newuniversalredisclient.err: %v", err)
		return err
	}
	a.cfgManager.WithRedisConnection(redisConn)

	return nil
}
