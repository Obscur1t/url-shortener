package url

import (
	"context"
	"errors"
	"log/slog"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/service"
	"url-shortener/internal/storage"
)

type PgStorage interface {
	SaveURL(ctx context.Context, url string, alias string, userID int) error
	GetURL(ctx context.Context, alias string) (string, error)
}

type Service struct {
	pgStorage PgStorage
	log       *slog.Logger
}

func New(pgStorage PgStorage, log *slog.Logger) *Service {
	return &Service{
		pgStorage: pgStorage,
		log:       log,
	}
}

func (s *Service) SaveURL(ctx context.Context, url string, userID int) (string, error) {
	for i := 0; i < 10; i++ {
		alias := random.NewRandomAlias(6)
		err := s.pgStorage.SaveURL(ctx, url, alias, userID)
		if err == nil {
			return alias, nil
		}
		if errors.Is(err, storage.ErrAlreadyExists) {
			s.log.Warn("server error", slog.String("Collision occurred, save url err", err.Error()))
			continue
		}
		if errors.Is(err, storage.ErrPostgres) {
			s.log.Error("server error", slog.String("Internal server error", err.Error()))
			return "", storage.ErrPostgres
		}
	}
	s.log.Error("service error", slog.String("save url", "attempts over"))
	return "", service.ErrAttemptsOver
}

func (s *Service) GetURL(ctx context.Context, alias string) (string, error) {
	return s.pgStorage.GetURL(ctx, alias)
}
