package repo

import (
	"context"

	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
)

type UserRepository interface {
	Create(ctx context.Context, username string, password string) (int, error)
	Authenticate(ctx context.Context, username string, password string) (int, error)
}
type OrderRepository interface {
	CreateNewOrder(ctx context.Context, order models.Order) error
	ListOrders(ctx context.Context, userID int) ([]*models.Order, error)
	ListWithdrawals(ctx context.Context, userID int) ([]*models.Order, error)

	CurrentBalance(ctx context.Context, userID int) (int, error)

	Withdraw(ctx context.Context, order models.Order) error

	ListUnprocessedOrders(ctx context.Context, limit, offset int) ([]*models.Order, error)
	UpdateOrder(ctx context.Context, order models.Order) error
}
