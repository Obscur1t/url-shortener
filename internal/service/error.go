package service

import (
	"errors"
)

var (
	ErrAttemptsOver   = errors.New("attempts over")
	ErrCreatePassHash = errors.New("failed to create password hash")
	ErrInvalidPass    = errors.New("invalid password")
)
