package url

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/middleware/auth"
	"url-shortener/internal/service"
	"url-shortener/internal/storage"
	"url-shortener/internal/transport/response"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Alias string `json:"alias"`
}

type Service interface {
	SaveURL(ctx context.Context, url string, userID int) (string, error)
	GetURL(ctx context.Context, alias string) (string, error)
}

type URLHandler struct {
	service Service
	log     *slog.Logger
}

func New(service Service, log *slog.Logger) *URLHandler {
	return &URLHandler{
		service: service,
		log:     log,
	}
}

func (h *URLHandler) SaveURL(w http.ResponseWriter, r *http.Request) {
	var request Request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		h.log.Error("handler error", slog.String("save url decode err", err.Error()))
		response.ResponseErr(h.log, w, http.StatusBadRequest, "Invalid request format. Check the transmitted data")
		return
	}
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		h.log.Error("handler error", slog.String("context", "user id not found or invalid type"))
		response.ResponseErr(h.log, w, http.StatusInternalServerError, "Internal server error")
		return
	}

	alias, err := h.service.SaveURL(r.Context(), request.URL, userID)
	if err != nil {
		if errors.Is(err, storage.ErrPostgres) {
			h.log.Error("handler error", slog.String("save url server error", err.Error()))
			response.ResponseErr(h.log, w, http.StatusInternalServerError, "Server error")
			return
		}
		if errors.Is(err, service.ErrAttemptsOver) {
			h.log.Error("handler error", slog.String("save url attempts over", err.Error()))
			response.ResponseErr(h.log, w, http.StatusInternalServerError, "Failed to create alias. Try again later")
			return
		}
		h.log.Error("handler error", slog.String("save url server error", err.Error()))
		response.ResponseErr(h.log, w, http.StatusInternalServerError, "Server error")
		return
	}

	var resp = Response{
		Alias: alias,
	}
	response.ResponseJSON(h.log, w, http.StatusCreated, resp)
}

func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")
	if alias == "" {
		h.log.Error("failed get alias", slog.String("error", "failed to get alias from url"))
		response.ResponseErr(h.log, w, http.StatusBadRequest, "failed to get alias from url")
		return
	}

	url, err := h.service.GetURL(r.Context(), alias)
	if err != nil {
		if errors.Is(err, storage.ErrPostgres) {
			h.log.Error("handler error", slog.String("redirect server error", err.Error()))
			response.ResponseErr(h.log, w, http.StatusInternalServerError, "Server error")
			return
		}
		if errors.Is(err, storage.ErrURLNotFound) {
			h.log.Error("handler error", slog.String("redirect error not found", err.Error()))
			response.ResponseErr(h.log, w, http.StatusNotFound, "Url not found. Check the transmitted data")
			return
		}
		h.log.Error("handler error", slog.String("save url server error", err.Error()))
		response.ResponseErr(h.log, w, http.StatusInternalServerError, "Server error")
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}
