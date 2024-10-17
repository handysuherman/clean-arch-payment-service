package infra

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	postgres "github.com/handysuherman/clean-arch-payment-service/internal/pkg/databases/postgresql"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
)

func (a *app) pqsql(ctx context.Context) error {
	psqlOpt := &postgres.Config{
		Host:      a.cfg.Databases.PostgreSQL.Host,
		Port:      a.cfg.Databases.PostgreSQL.Port,
		User:      a.cfg.Databases.PostgreSQL.Username,
		DBName:    a.cfg.Databases.PostgreSQL.DBName,
		Password:  a.cfg.Databases.PostgreSQL.Password,
		EnableTls: a.cfg.Databases.PostgreSQL.EnableTLS,
	}

	if psqlOpt.EnableTls {
		pqsqlTls, err := helper.Base64EncodedTLS(a.cfg.TLS.PostgreSQL.Ca, a.cfg.TLS.PostgreSQL.Cert, a.cfg.TLS.PostgreSQL.Key)
		if err != nil {
			a.log.Warnf("pqsql.helper.base64encodedtls.err: %v", err)
			return err
		}

		pqsqlTls.ServerName = a.cfg.Databases.PostgreSQL.Host
		psqlOpt.TLs = pqsqlTls
	}

	pgxConn, err := postgres.NewPgxConn(ctx, psqlOpt)
	if err != nil {
		a.log.Warnf("pqsql.postgres.newpgxconn.err: %v", err)
		return err
	}
	a.cfgManager.WithPqsqlConnection(pgxConn)

	return nil
}

func (a *app) runDBMigration() error {
	source := a.cfg.Databases.PostgreSQL.Source

	if a.cfg.Databases.PostgreSQL.EnableTLS {
		source = fmt.Sprintf("%v&sslrootcert=%v&sslkey=%v&sslcert=%v", a.cfg.Databases.PostgreSQL.TLsSource, a.cfg.Databases.PostgreSQL.Capath, a.cfg.Databases.PostgreSQL.Keypath, a.cfg.Databases.PostgreSQL.Certpath)
	}

	migration, err := migrate.New(a.cfg.Databases.PostgreSQL.MigrationURL, source)
	if err != nil {
		a.log.Warnf("runDBMigration.migrate.new.err: %v", err)
		return err
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		a.log.Warnf("runDBMigration.migration.up.err: %v", err)
		return err
	}

	a.log.Info("DB migrated successfully...")

	return nil
}
