package postgres

import "errors"

var (
	ErrCreatePool = errors.New("failed to create pool")
	ErrPingPool   = errors.New("failed to ping pool")
)
