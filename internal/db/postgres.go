package db

import (
	"context"
	"embed"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations
var migrationsFS embed.FS

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

	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		log.Error("failed to create migration source", "error", err)
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
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
