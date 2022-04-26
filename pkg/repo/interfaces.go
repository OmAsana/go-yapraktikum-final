package repo

import (
	"context"
	"errors"

	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
)

type Error error

var (
	ErrUserNotFound      Error = errors.New("user does not exist")
	ErrUserAuthFailed    Error = errors.New("user authentication failed")
	ErrUserAlreadyExists Error = errors.New("duplicate user name")

	ErrOrderAlreadyUploadedByCurrentUser Error = errors.New("order already exist for this user")
	ErrOrderCreatedByAnotherUser         Error = errors.New("order already exist for another user")

	ErrInternalError Error = errors.New("internal error")
)

type User interface {
	Create(ctx context.Context, username string, password string) (int, Error)
	Authenticate(ctx context.Context, username string, password string) (int, Error)
}
type Order interface {
	CreateNewOrder(ctx context.Context, order models.Order) Error
	ListOrders(ctx context.Context, userID int) ([]*models.Order, Error)
	ListWithdrawals(ctx context.Context, userID int) ([]*models.Order, Error)
	CurrentBalance(ctx context.Context, userID int) (int, Error)

	ListUnprocessedOrders(ctx context.Context, limit, offset int) ([]*models.Order, Error)
	UpdateOrder(ctx context.Context, order models.Order) Error
}
