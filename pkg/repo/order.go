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

//func (u *orderRepo) CreateNewOrder(ctx context.Context, order models.Order) Error {
//	var err error
//	l := u.log.With(zap.Int("order", order.OrderID))
//	defer func() {
//		if err != nil {
//			l.Error("error creating order", zap.Error(err))
//		}
//	}()
//	sqlStatement := `INSERT INTO orders (order_id, status, tx_type, accrual, user_id, uploaded_at)
//VALUES ($1, $2, $3, $4, $5, $6)`
//
//	res, err := u.db.ExecContext(ctx, sqlStatement,
//		order.OrderID,
//		models.NewStatus,
//		order.TXType,
//		order.Accrual,
//		order.UserID,
//		time.Now())
//	if err != nil {
//		// duplicate key error
//		if strings.Contains(err.Error(), "SQLSTATE 23505") {
//			return ErrDuplicateOrder
//		}
//	}
//
//	inserts, err := res.RowsAffected()
//	if err != nil {
//		return ErrInternalError
//
//	}
//
//	if inserts != 1 {
//		// TODO
//		// Implement  zapcore.ObjectMarshaler
//		u.log.Debug("Did not create order", zap.Reflect("order", order))
//		return ErrInternalError
//	}
//
//	return nil
//}

func (u *orderRepo) CreateNewOrder(ctx context.Context, order models.Order) Error {
	l := u.log.With(zap.Int("order", order.OrderID))

	findOrderSql := `SELECT user_id FROM orders WHERE order_id = $1`
	var userID int
	err := u.db.QueryRowContext(ctx, findOrderSql, order.OrderID).Scan(&userID)
	switch {
	// Create new order if it does not exist in the db
	case err == sql.ErrNoRows:
		sqlStatement := `INSERT INTO orders (order_id, status, tx_type, accrual, user_id, uploaded_at)
	VALUES ($1, $2, $3, $4, $5, $6)`

		res, err := u.db.ExecContext(ctx, sqlStatement,
			order.OrderID,
			models.NewStatus,
			order.TXType,
			order.Accrual,
			order.UserID,
			time.Now())
		if err != nil {
			// duplicate key error
			if strings.Contains(err.Error(), "SQLSTATE 23505") {
				return ErrDuplicateOrder
			}
		}

		inserts, err := res.RowsAffected()
		if err != nil {
			return ErrInternalError
		}

		if inserts != 1 {
			// TODO
			// Implement  zapcore.ObjectMarshaler
			l.Debug("Did not create order", zap.Reflect("order", order))
			return ErrInternalError
		}
		return nil
	case err != nil:
		return ErrInternalError
	}

	if userID == order.UserID {
		return ErrOrderAlreadyUploadedByCurrentUser
	} else {
		return ErrOrderCreatedByAnotherUser
	}

}

func (u *orderRepo) ListOrders(ctx context.Context, userID int) ([]*models.Order, error) {
	return u.queryOrders(ctx, userID, models.DepositOrder)
}

func (u *orderRepo) ListWithdrawals(ctx context.Context, userID int) ([]*models.Order, error) {
	return u.queryOrders(ctx, userID, models.WithdrawalOrder)
}

func (u *orderRepo) queryOrders(ctx context.Context, userID int, orderType models.OrderType) ([]*models.Order, error) {
	var err error
	l := logr.FromContext(ctx)
	defer func() {
		if err != nil {
			l.Error("error getting user orders", zap.Error(err))
		}
	}()

	sqlStatement := `SELECT order_id, status, tx_type, accrual, user_id, uploaded_at, processed_at
FROM orders 
WHERE user_id = $1 AND tx_type = $2`

	rows, err := u.db.QueryContext(ctx, sqlStatement, userID, orderType)
	if err != nil {
		return nil, ErrInternalError
	}
	err = rows.Err()
	if err != nil {
		return nil, ErrInternalError
	}

	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		var t sql.NullTime

		err = rows.Scan(
			&order.OrderID,
			&order.Status,
			&order.TXType,
			&order.Accrual,
			&order.UserID,
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

func (u *orderRepo) CurrentBalance(ctx context.Context, userID int) (int, Error) {
	var err error
	l := logr.FromContext(ctx)
	defer func() {
		if err != nil {
			l.Error("error getting user balance", zap.Error(err))
		}
	}()

	sqlStatement := `SELECT COALESCE(SUM(accrual),0) AS total FROM orders WHERE user_id = $1`

	var sum int
	err = u.db.QueryRowContext(ctx, sqlStatement, userID).Scan(&sum)
	if err != nil {
		return -1, ErrInternalError
	}
	return sum, nil
}
