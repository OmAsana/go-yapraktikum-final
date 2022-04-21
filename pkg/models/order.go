package models

import "time"

type OrderType string

var (
	DepositOrder    OrderType = "deposit"
	WithdrawalOrder OrderType = "withdrawal"
)

type OrderStatus string

var (
	NewStatus        OrderStatus = "NEW"
	ProcessingStatus OrderStatus = "PROCESSING"

	InvalidStatus   OrderStatus = "INVALID"
	ProcessedStatus OrderStatus = "PROCESSED"
)

type Order struct {
	OrderId     int
	Status      OrderStatus
	TXType      OrderType
	Accrual     int
	UserId      int
	UploadedAt  time.Time
	ProcessedAt time.Time
}
