package postgres

import (
	"context"
	"errors"
	"log/slog"
	"url-shortener/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UrlRepo struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewUrlRepo(pool *pgxpool.Pool, log *slog.Logger) *UrlRepo {
	return &UrlRepo{
		pool: pool,
		log:  log,
	}
}

func (r *UrlRepo) SaveURL(ctx context.Context, url string, alias string, userID int) error {
	query := "INSERT INTO urls(url, alias, user_id) VALUES($1, $2, $3)"
	_, err := r.pool.Exec(ctx, query, url, alias, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				r.log.Error("Postgres error", slog.String("url already exists", err.Error()))
				return storage.ErrAlreadyExists
			}
		}
		r.log.Error("Postgres error", slog.String("postgres error", err.Error()))
		return storage.ErrPostgres
	}

	return nil
}

func (r *UrlRepo) GetURL(ctx context.Context, alias string) (string, error) {
	query := "SELECT url FROM urls WHERE alias=$1"
	var url string
	err := r.pool.QueryRow(ctx, query, alias).Scan(&url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Error("Postgres error", slog.String("url not found", err.Error()))
			return "", storage.ErrURLNotFound
		}
		r.log.Error("Postgres error", slog.String("postgres error", err.Error()))
		return "", storage.ErrPostgres
	}

	return url, nil
}
