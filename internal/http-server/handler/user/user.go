package user

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

type Service interface {
	Registration(ctx context.Context, email, password string) (int, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type UserHandler struct {
	service Service
	log     *slog.Logger
}

func New(service Service, log *slog.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		log:     log,
	}
}

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ResponseRegistration struct {
	ID int `json:"id"`
}

type ResponseLogin struct {
	Token string `json:"token"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.log.Error("handler error", slog.Any("Registration error", err))
		response.ResponseErr(h.log, w, http.StatusBadRequest, "server error")
		return
	}
	if request.Email == "" || request.Password == "" {
		h.log.Error("handler error", slog.Any("Registration error", "invalid password or email"))
		response.ResponseErr(h.log, w, http.StatusBadRequest, "invalid password or email")
		return
	}

	id, err := h.service.Registration(r.Context(), request.Email, request.Password)
	if err != nil {
		h.log.Error("handler error", slog.Any("Registration error", err))
		if errors.Is(err, service.ErrCreatePassHash) {
			response.ResponseErr(h.log, w, http.StatusInternalServerError, "server error")
			return
		}
		if errors.Is(err, storage.ErrAlreadyExists) {
			response.ResponseErr(h.log, w, http.StatusConflict, "email already exists")
			return
		}
		response.ResponseErr(h.log, w, http.StatusInternalServerError, "server error")
		return
	}
	resp := ResponseRegistration{
		ID: id,
	}
	response.ResponseJSON(h.log, w, http.StatusCreated, resp)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.log.Error("handler error", slog.Any("Login error", err))
		response.ResponseErr(h.log, w, http.StatusBadRequest, "server error")
		return
	}
	if request.Email == "" || request.Password == "" {
		h.log.Error("handler error", slog.Any("Login error", "invalid password or email"))
		response.ResponseErr(h.log, w, http.StatusBadRequest, "invalid password or email")
		return
	}

	token, err := h.service.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		h.log.Error("handler error", slog.Any("Login error", err))
		if errors.Is(err, storage.ErrUserNotFound) {
			response.ResponseErr(h.log, w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		if errors.Is(err, service.ErrInvalidPass) {
			response.ResponseErr(h.log, w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		response.ResponseErr(h.log, w, http.StatusInternalServerError, "server error")
		return
	}
	resp := ResponseLogin{
		Token: token,
	}
	response.ResponseJSON(h.log, w, http.StatusOK, resp)
}
