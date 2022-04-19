package repo

import (
	"context"
	"database/sql"

	"go.uber.org/zap"

	logr "github.com/OmAsana/go-yapraktikum-final/pkg/logger"
	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
)

var _ Order = (*orderRepo)(nil)

type orderRepo struct {
	db  *sql.DB
	log *zap.Logger
}

func newOrderRepo(db *sql.DB, logger *zap.Logger) *orderRepo {
	if logger == nil {
		logger = logr.NewNoop()
	}
	return &orderRepo{db: db, log: logger}
}

func (u orderRepo) CreateNewOrder(ctx context.Context, order models.Order) Error {
	//TODO implement me
	panic("implement me")
}

func (u orderRepo) ListOrders(ctx context.Context, username string) []*models.Order {
	//TODO implement me
	panic("implement me")
}

func (u orderRepo) ListWithdrawals(ctx context.Context, username string) []*models.Order {
	//TODO implement me
	panic("implement me")
}
