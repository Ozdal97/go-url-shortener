package domain

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrExpired          = errors.New("expired")
)
