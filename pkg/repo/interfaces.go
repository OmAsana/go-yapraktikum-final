package repo

import (
	"context"
	"errors"

	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
)

type Error error

var (
	UserNotFound   Error = errors.New("user does not exist")
	UserAuthFailed Error = errors.New("user authentication failed")

	DuplicateOrder Error = errors.New("duplicate order")

	InternalError Error = errors.New("internal error")
)

type User interface {
	Create(ctx context.Context, username string, pwdHash string) Error
	Authenticate(ctx context.Context, username string, pwdHash string) Error
	CurrentBalance(ctx context.Context, username string) int
}

type Order interface {
	CreateNewOrder(ctx context.Context, order models.Order) Error
	ListOrders(ctx context.Context, username string) []*models.Order
	ListWithdrawals(ctx context.Context, username string) []*models.Order
}
