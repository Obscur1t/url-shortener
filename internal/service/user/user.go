package user

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"url-shortener/internal/lib/jwt"
	"url-shortener/internal/service"
	"url-shortener/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Email    string
	PassHash string
}

type UserStorage interface {
	Registration(ctx context.Context, email, passwordHash string) (int, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type UserService struct {
	storage  UserStorage
	log      *slog.Logger
	secret   string
	tokenTTL time.Duration
}

func New(storage UserStorage, log *slog.Logger, secret string, tokenTTL time.Duration) *UserService {
	return &UserService{storage: storage, log: log, secret: secret, tokenTTL: tokenTTL}
}

func (u *UserService) Registration(ctx context.Context, email, password string) (int, error) {
	passwordByte := []byte(password)
	passwordHash, err := bcrypt.GenerateFromPassword(passwordByte, bcrypt.DefaultCost)
	if err != nil {
		u.log.Error("service error", slog.Any("registration", err))
		return 0, service.ErrCreatePassHash
	}

	return u.storage.Registration(ctx, email, string(passwordHash))
}

func (u *UserService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.storage.GetByEmail(ctx, email)
	if err != nil {
		u.log.Error("service error", slog.Any("login", err))
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", storage.ErrUserNotFound
		}
		return "", storage.ErrPostgres
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password))
	if err != nil {
		return "", service.ErrInvalidPass
	}
	return jwt.NewToken(u.secret, user.ID, u.tokenTTL)
}
