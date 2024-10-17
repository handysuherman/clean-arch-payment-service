package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/handysuherman/clean-arch-payment-service/internal/config"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/tracing"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
)

func (r *Store) execTx(ctx context.Context, fnCallback func(*Queries) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fnCallback(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %w, rb err: %w", err, rbErr)
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func findModuleRoot(dir string) string {
	for {
		_, err := os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

func (r *Store) OnConfigUpdate(key string, config *config.App) {
	r.log.Infof("received update from '%s' key", key)

	r.cfg = config

	r.log.Infof("updated configuration from '%s' key successfully applied", key)
}

func (r *Store) OnPqsqlUpdate(key string, pqsqlConnection *pgxpool.Pool) {
	r.log.Infof("receive a new connection from newly updated key: %v", key)

	r.db = pqsqlConnection

	r.Queries = New(r.db)

	r.log.Info("newly updated pqsql connection successfully applied")
}

func (r *Store) OnRedisUpdate(key string, redisConnection redis.UniversalClient) {
	r.log.Infof("receive a new connection from newly updated key: %v", key)

	// r.RedisRepositoryImpl = NewRedisRepositoryImpl(r.log, r.cfg, redisConnection)

	r.log.Info("newly updated redis connection successfully applied")
}

func errorResponse(
	span opentracing.Span,
	err error,
	fullError error,
	returnedFormat string,
	loggedFormat string,
) error {
	tracing.TraceWithError(span, fmt.Errorf("%s: %v", loggedFormat, fullError))
	return fmt.Errorf("%s: %v", returnedFormat, err)
}
