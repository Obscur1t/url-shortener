package storage

import "errors"

var (
	ErrAlreadyExists = errors.New("url already exists")
	ErrPostgres      = errors.New("postgres error")
	ErrURLNotFound   = errors.New("url not found")
	ErrUserNotFound  = errors.New("user not found ")
)
