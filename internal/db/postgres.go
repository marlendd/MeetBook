package db

import (
	"context"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, dsn string, log *slog.Logger) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Error("failed to connect to db", "error", err)
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		log.Error("failed to ping the db", "error", err)
		return nil, err
	}

	return pool, nil
}

func RunMigrations(dsn string, log *slog.Logger) error {
	log.Debug("running migration")
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.Error("failed to run migrations", "error", err)
		return err
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info("no migrations to apply")
			return nil
		}
		log.Error("failed to run migrations", "error", err)
		return err
	}

	log.Debug("migrations applied successfully")
	return nil
}
