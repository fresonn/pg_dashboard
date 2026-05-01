package postgres

import "errors"

var ErrNotConnected = errors.New("postgres connection not initialized")
