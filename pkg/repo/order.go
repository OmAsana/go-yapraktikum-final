package repo

import (
	"context"
	"database/sql"
	"strings"
	"time"

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

func (u *orderRepo) CreateNewOrder(ctx context.Context, order models.Order) Error {
	var err error
	l := u.log.With(zap.Int("order", order.OrderId))
	defer func() {
		if err != nil {
			l.Error("error creating order", zap.Error(err))
		}
	}()
	sqlStatement := `
INSERT INTO orders (order_id, status, tx_type, accrual, user_id, uploaded_at) 
VALUES ($1, $2, $3, $4, $5, $6)`

	res, err := u.db.ExecContext(ctx, sqlStatement,
		order.OrderId,
		models.NewStatus,
		order.TXType,
		order.Accrual,
		order.UserId,
		time.Now())
	if err != nil {
		// duplicate key error
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			return DuplicateOrder
		}
	}

	inserts, err := res.RowsAffected()
	if err != nil {
		return InternalError

	}

	if inserts != 1 {
		// TODO
		// Implement  zapcore.ObjectMarshaler
		u.log.Debug("Did not create order", zap.Reflect("order", order))
		return InternalError
	}

	return nil
}

func (u *orderRepo) ListOrders(ctx context.Context, userId int) ([]*models.Order, error) {
	return u.queryOrders(ctx, userId, models.DepositOrder)
}

func (u *orderRepo) ListWithdrawals(ctx context.Context, userId int) ([]*models.Order, error) {
	return u.queryOrders(ctx, userId, models.WithdrawalOrder)
}

func (u *orderRepo) queryOrders(ctx context.Context, userId int, orderType models.OrderType) ([]*models.Order, error) {
	var err error
	l := u.log.With(zap.Int("user_id", userId))
	defer func() {
		if err != nil {
			l.Error("error getting user orders", zap.Error(err))
		}
	}()

	sqlStatement := `
SELECT order_id, status, tx_type, accrual, user_id, uploaded_at, processed_at
FROM orders 
WHERE user_id = $1 AND tx_type = $2
`

	rows, err := u.db.QueryContext(ctx, sqlStatement, userId, orderType)
	if err != nil {
		return nil, InternalError
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		var t sql.NullTime

		err = rows.Scan(
			&order.OrderId,
			&order.Status,
			&order.TXType,
			&order.Accrual,
			&order.UserId,
			&order.UploadedAt,
			&t,
		)

		if t.Valid {
			order.ProcessedAt = t.Time
		}

		if err != nil {
			return nil, err
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

func (u *orderRepo) CurrentBalance(ctx context.Context, userId int) (int, Error) {
	var err error
	l := u.log.With(zap.Int("user_id", userId))
	defer func() {
		if err != nil {
			l.Error("error getting user balance", zap.Error(err))
		}
	}()

	sqlStatement := `SELECT COALESCE(SUM(accrual),0) AS total FROM orders WHERE user_id = $1`

	var sum int
	err = u.db.QueryRowContext(ctx, sqlStatement, userId).Scan(&sum)
	if err != nil {
		return -1, InternalError
	}
	return sum, nil
}
