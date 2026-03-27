package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"url-shortener/internal/lib/jwt"
	"url-shortener/internal/transport/response"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(log *slog.Logger, secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				log.Warn("auth middleware", slog.String("auth token", "empty auth token"))
				response.ResponseErr(log, w, http.StatusUnauthorized, "invalid token")
				return
			}
			if !strings.HasPrefix(auth, "Bearer ") {
				log.Warn("auth middleware", slog.String("auth token", "invalid header "))
				response.ResponseErr(log, w, http.StatusUnauthorized, "invalid token")
				return
			}
			token := strings.TrimPrefix(auth, "Bearer ")
			id, err := jwt.ParseToken(token, secret)
			if err != nil {
				log.Warn("auth middleware", slog.Any("auth token", err))
				response.ResponseErr(log, w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
