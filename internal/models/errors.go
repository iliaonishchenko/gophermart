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

	// Accrual errors
	ErrAccrualClientOrderNotRegistered = errors.New("accrual order not registered")
	ErrAccrualClientTooManyRequests    = errors.New("accrual order too many requests")
	ErrAccrualClientInternalError      = errors.New("accrual order internal error")
)
