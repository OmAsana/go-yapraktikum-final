package models

import (
	"time"

	"github.com/theplant/luhn"
)

type OrderType string

var (
	DepositOrder    OrderType = "deposit"
	WithdrawalOrder OrderType = "withdrawal"
)

type OrderStatus string

var (
	NewStatus        OrderStatus = "NEW"
	ProcessingStatus OrderStatus = "PROCESSING"
	InvalidStatus    OrderStatus = "INVALID"
	ProcessedStatus  OrderStatus = "PROCESSED"
)

type Order struct {
	OrderID     int
	Status      OrderStatus
	TXType      OrderType
	Accrual     float64
	UserID      int
	UploadedAt  time.Time
	ProcessedAt time.Time
}

func NewOrder(orderID int, userID int) Order {
	return Order{
		OrderID: orderID,
		Status:  NewStatus,
		TXType:  DepositOrder,
		Accrual: 0,
		UserID:  userID,
	}
}

func (o Order) Valid() bool {
	return luhn.Valid(o.OrderID)
}
