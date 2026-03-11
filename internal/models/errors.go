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

	// Withdrawals errors
	ErrWithdrawalNonExistentOrder = errors.New("withdrawal relates to non-existing order")
	ErrWithdrawalNotEnoughFunds   = errors.New("withdrawal not enough funds")

	// Auth/context errors
	ErrNoUserInContext   = errors.New("no user UUID in context")
	ErrInvalidUserType   = errors.New("user UUID in context has unexpected type")

	// Accrual errors
	ErrAccrualClientOrderNotRegistered = errors.New("accrual order not registered")
)
