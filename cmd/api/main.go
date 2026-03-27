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
	"url-shortener/internal/http-server/handler/url"
	"url-shortener/internal/http-server/handler/user"
	"url-shortener/internal/infrastructure/postgres"
	"url-shortener/internal/logger"
	"url-shortener/internal/middleware/auth"
	loggerMiddleware "url-shortener/internal/middleware/logger"
	"url-shortener/internal/middleware/recovery"
	urlsrv "url-shortener/internal/service/url"
	usersrv "url-shortener/internal/service/user"
	postgresRepo "url-shortener/internal/storage/postgres"
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

	urlRepo := postgresRepo.NewUrlRepo(pool, logger)
	userRepo := postgresRepo.NewUserRepo(pool, logger)

	urlSrv := urlsrv.New(urlRepo, logger)
	userSrv := usersrv.New(userRepo, logger, cfg.Auth.Secret, cfg.Auth.TokenTTL)

	urlHand := url.New(urlSrv, logger)
	userHand := user.New(userSrv, logger)

	mux := http.NewServeMux()
	router := recovery.RecoveryMiddleware(logger, loggerMiddleware.LoggerMiddleware(logger, mux))

	authMw := auth.AuthMiddleware(logger, cfg.Auth.Secret)

	mux.HandleFunc("POST /register", userHand.Register)
	mux.HandleFunc("POST /login", userHand.Login)
	mux.HandleFunc("GET /url/{alias}", urlHand.RedirectURL)
	mux.Handle("POST /url", authMw(http.HandlerFunc(urlHand.SaveURL)))

	server := http.Server{
		Addr:         cfg.App.AppAddr,
		Handler:      router,
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
