package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handler"
	"url-shortener/internal/infrastructure/postgres"
	"url-shortener/internal/logger"
	"url-shortener/internal/service"
	pgRepo "url-shortener/internal/storage/postgres"
)

func main() {
	cfg := config.MustLoad()
	logger := logger.SetupLogger(cfg.App.AppEnv)
	dsn := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.DBName,
		cfg.DB.SSL,
	)
	pool, err := postgres.New(context.Background(), logger, dsn)
	if err != nil {
		logger.Error("failed to init storage", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("Database connection established")
	logger.Info("Starting URL Shortener", slog.String("env", cfg.App.AppEnv))

	pgRepo := pgRepo.New(pool, logger)
	service := service.New(pgRepo, logger)
	handler := handler.New(service, logger)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /url", handler.SaveURL)
	mux.HandleFunc("GET /url/{alias}", handler.RedirectURL)

	server := http.Server{
		Addr:         cfg.App.AppAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to server", slog.String("Error", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	logger.Info("Stopping server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server.Shutdown(ctx)
}
