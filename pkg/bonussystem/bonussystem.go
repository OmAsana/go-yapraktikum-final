package bonussystem

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/go-resty/resty/v2"

	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
	"github.com/OmAsana/go-yapraktikum-final/pkg/repo"
)

type OrderStatus string

var (
	StatusRegistered OrderStatus = "REGISTERED"
	StatusInvalid    OrderStatus = "INVALID"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusProcessed  OrderStatus = "PROCESSED"
)

type AccrualResp struct {
	Order   string
	Status  OrderStatus
	Accrual float64
}

type BonusSystem struct {
	endpoint  string
	orderRepo repo.OrderRepository
	log       *zap.Logger
	client    *resty.Client
}

func NewBonusSystem(endpoint string, orderRepo repo.OrderRepository, logger *zap.Logger) *BonusSystem {
	client := resty.New()
	client.SetBaseURL(endpoint + "/api/orders")

	return &BonusSystem{
		orderRepo: orderRepo,
		log:       logger,
		client:    client,
	}
}

func (s *BonusSystem) updateOrders(ctx context.Context, orders []*models.Order) {

	for _, o := range orders {
		resp, err := s.client.R().SetContext(ctx).Get(fmt.Sprintf("%d", o.OrderID))
		if err != nil {
			s.log.Error("Error quering remote api", zap.Error(err))
			return
		}
		if resp.IsError() {
			s.log.Error("Remote api return error", zap.Any("err", resp.Error()))
			return
		}

		var accrualResp AccrualResp

		err = json.Unmarshal(resp.Body(), &accrualResp)
		if err != nil {
			s.log.Error("Error unmarshaling json", zap.Error(err))
			return
		}

		switch accrualResp.Status {
		case StatusInvalid:
			o.Status = models.InvalidStatus
		case StatusProcessed:
			o.Status = models.ProcessedStatus
			o.Accrual = accrualResp.Accrual
		default:
			// do no process other statuses
			continue
		}

		o.ProcessedAt = time.Now()

		err = s.orderRepo.UpdateOrder(ctx, *o)
		if err != nil {
			s.log.Error("Failed to update order", zap.Error(err))
		}
	}
}

func (s *BonusSystem) processOrders(ctx context.Context) {
	limit := 10
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			orders, err := s.orderRepo.ListUnprocessedOrders(ctx, limit, offset)
			if err != nil {
				s.log.Error("Error fetching orders", zap.Error(err))
				return
			}
			if len(orders) == 0 {
				s.log.Info("No orders to process")
				return
			}
			s.log.Info(fmt.Sprintf("Processing %d orders", len(orders)))
			s.updateOrders(ctx, orders)
			offset = offset + limit
		}
	}
}

func (s *BonusSystem) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			s.log.Info("Shutting down")
			return nil
		case <-time.After(1 * time.Second):
			s.processOrders(ctx)
		}
	}
}
