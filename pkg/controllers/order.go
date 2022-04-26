package controllers

import (
	"strconv"
	"time"

	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
)

type Order struct {
	Number     string             `json:"number"`
	Status     models.OrderStatus `json:"status"`
	Accrual    int                `json:"accrual,omitempty"`
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
