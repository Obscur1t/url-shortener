package postgres

import (
	"context"
	"errors"
	"log/slog"
	"url-shortener/internal/service/user"
	"url-shortener/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewUserRepo(pool *pgxpool.Pool, log *slog.Logger) *UserRepo {
	return &UserRepo{
		pool: pool,
		log:  log,
	}
}

func (u *UserRepo) Registration(ctx context.Context, email, passwordHash string) (int, error) {
	query := "INSERT INTO users(email, password_hash) VALUES($1, $2) RETURNING id"

	var id int

	err := u.pool.QueryRow(ctx, query, email, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				u.log.Error("Postgres error", slog.Any("email already exists", err))
				return 0, storage.ErrAlreadyExists
			}
		}
		u.log.Error("Postgres error", slog.Any("registration error", err))
		return 0, storage.ErrPostgres
	}

	return id, nil
}

func (u *UserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := "SELECT id, email, password_hash FROM users WHERE email=$1"

	var user user.User
	err := u.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			u.log.Error("Postgres error", slog.Any("email not found ", err))
			return nil, storage.ErrUserNotFound
		}
		u.log.Error("Postgres error", slog.Any("get by email", err))
		return nil, storage.ErrPostgres
	}
	return &user, nil
}
