package service

import "errors"

var (
	ErrUnauthorized        = errors.New("unauthorized")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrInvalidBuilding     = errors.New("invalid building")
	ErrInvalidNotifyTime   = errors.New("invalid notify time")
	ErrUserQueryFailed     = errors.New("user query failed")
	ErrUserCreateFailed    = errors.New("user create failed")
	ErrTokenGenerateFailed = errors.New("token generate failed")
)
