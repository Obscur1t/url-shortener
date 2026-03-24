package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
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
	SaveUrl(ctx context.Context, url string) (string, error)
	GetUrl(ctx context.Context, alias string) (string, error)
}

type Handler struct {
	service Service
	log     *slog.Logger
}

func New(service Service, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) SaveURL(w http.ResponseWriter, r *http.Request) {
	var request Request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		h.log.Error("handler error", slog.String("save url decode err", err.Error()))
		response.ResponseErr(h.log, w, http.StatusBadRequest, "Invalid request format. Check the transmitted data")
	}

	alias, err := h.service.SaveUrl(r.Context(), request.URL)
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

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")
	if alias == "" {
		h.log.Error("failed get alias", slog.String("error", "failed to get alias from url"))
		response.ResponseErr(h.log, w, http.StatusBadRequest, "failed to get alias from url")
		return
	}

	url, err := h.service.GetUrl(r.Context(), alias)
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
