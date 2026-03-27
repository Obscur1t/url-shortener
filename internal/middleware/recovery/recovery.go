package recovery

import (
	"log/slog"
	"net/http"
	"url-shortener/internal/transport/response"
)

func RecoveryMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered", slog.Any("error", err))
				response.ResponseErr(log, w, http.StatusInternalServerError, "server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
