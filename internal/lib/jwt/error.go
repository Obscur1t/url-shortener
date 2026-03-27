package jwt

import "errors"

var (
	ErrMethod       = errors.New("invalid method")
	ErrParseToken   = errors.New("failed to parse token")
	ErrInvalidToken = errors.New("invalid token")
)
