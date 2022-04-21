package repo

import (
	"context"
	"errors"

	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
)

type Error error

var (
	UserNotFound      Error = errors.New("user does not exist")
	UserAuthFailed    Error = errors.New("user authentication failed")
	UserAlreadyExists Error = errors.New("duplicate user name")

	DuplicateOrder Error = errors.New("duplicate order")

	InternalError Error = errors.New("internal error")
)

type User interface {
	Create(ctx context.Context, username string, pwdHash string) Error
	Authenticate(ctx context.Context, username string, pwdHash string) (int, Error)
}
type Order interface {
	CreateNewOrder(ctx context.Context, order models.Order) Error
	ListOrders(ctx context.Context, userId int) ([]*models.Order, error)
	ListWithdrawals(ctx context.Context, userId int) ([]*models.Order, error)
	CurrentBalance(ctx context.Context, userId int) (int, Error)
}
