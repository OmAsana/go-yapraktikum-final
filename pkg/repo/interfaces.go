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

	ErrDuplicateOrder Error = errors.New("duplicate order")

	ErrInternalError Error = errors.New("internal error")
)

type User interface {
	Create(ctx context.Context, username string, pwdHash string) Error
	Authenticate(ctx context.Context, username string, pwdHash string) (int, Error)
}
type Order interface {
	CreateNewOrder(ctx context.Context, order models.Order) Error
	ListOrders(ctx context.Context, userID int) ([]*models.Order, error)
	ListWithdrawals(ctx context.Context, userID int) ([]*models.Order, error)
	CurrentBalance(ctx context.Context, userID int) (int, Error)
}
