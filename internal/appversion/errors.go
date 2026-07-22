package appversion

import "errors"

var (
	ErrNotFound = errors.New("app version not found")
	ErrConflict = errors.New("app version conflict")
)
