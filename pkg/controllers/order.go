package controllers

import (
	"strconv"
	"time"

	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
)

type Order struct {
	Number     string             `json:"number"`
	Status     models.OrderStatus `json:"status"`
	Accrual    float64            `json:"accrual,omitempty"`
	UploadedAt string             `json:"uploaded_at"`
}

func OrderModelToController(mo models.Order) Order {
	o := Order{
		Number:     strconv.Itoa(mo.OrderID),
		Status:     mo.Status,
		UploadedAt: mo.UploadedAt.Format(time.RFC3339),
	}

	if mo.Status != models.InvalidStatus {
		o.Accrual = mo.Accrual
	}

	return o
}

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (w Withdrawal) ToOrder(userID int) (models.Order, error) {
	orderID, err := strconv.Atoi(w.Order)
	if err != nil {
		return models.Order{}, err
	}
	return models.Order{
		OrderID:    orderID,
		Status:     models.NewStatus,
		TXType:     models.WithdrawalOrder,
		Accrual:    w.Sum,
		UserID:     userID,
		UploadedAt: time.Now()}, nil
}
