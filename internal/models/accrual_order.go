package models

type AccrualOrder struct {
	Order   string             `json:"order"`
	Status  AccrualOrderStatus `json:"status"`
	Accrual *float32           `json:"accrual"`
}

type AccrualOrderStatus string

const (
	StatusRegistered AccrualOrderStatus = "REGISTERED"
	StatusInvalid    AccrualOrderStatus = "INVALID"
	StatusProcessing AccrualOrderStatus = "PROCESSING"
	StatusProcessed  AccrualOrderStatus = "PROCESSED"
)
