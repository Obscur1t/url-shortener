package postgres

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, log *slog.Logger, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Info("Create pool", slog.String("failed to create pool", err.Error()))
		return nil, ErrCreatePool
	}

	if err := pool.Ping(ctx); err != nil {
		log.Info("Ping pool", slog.String("failed to ping pool", err.Error()))
		return nil, ErrPingPool
	}

	return pool, nil
}
