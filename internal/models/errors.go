package models

import "errors"

var (
	// User errors
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Order errors
	ErrOrderAlreadyExists       = errors.New("order already exists")
	ErrOrderExistsDifferentUser = errors.New("order already exists by another user")
	ErrInvalidOrderFormat       = errors.New("invalid order format")
)
