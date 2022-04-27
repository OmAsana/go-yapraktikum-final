package repo

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"

	logr "github.com/OmAsana/go-yapraktikum-final/pkg/logger"
	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
)

var _ OrderRepository = (*orderRepo)(nil)

type orderRepo struct {
	db  *sql.DB
	log *zap.Logger
}

func (u *orderRepo) Withdraw(ctx context.Context, order models.Order) error {
	l := logr.FromContext(ctx)

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		l.Error("Could not begin tx", zap.Error(err))
	}
	defer tx.Rollback()

	sqlStatement := `SELECT COALESCE(SUM(accrual),0) AS total FROM orders WHERE user_id = $1 and tx_type = $2`

	var depositSum float64
	err = tx.QueryRowContext(ctx, sqlStatement, order.UserID, models.DepositOrder).Scan(&depositSum)
	if err != nil {
		l.Error("Error querying db", zap.Error(err))
		return ErrInternalError
	}

	var withdrawSum float64
	err = tx.QueryRowContext(ctx, sqlStatement, order.UserID, models.WithdrawalOrder).Scan(&withdrawSum)
	if err != nil {
		l.Error("Error querying db", zap.Error(err))
		return ErrInternalError
	}

	if depositSum-withdrawSum > order.Accrual {
		sqlStatement := `INSERT INTO orders (order_id, status, tx_type, accrual, user_id, uploaded_at, processed_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`
		_, err := tx.ExecContext(ctx, sqlStatement,
			order.OrderID,
			models.ProcessedStatus,
			order.TXType,
			order.Accrual,
			order.UserID,
			order.UploadedAt,
			time.Now())
		if err != nil {
			l.Error("Error processing withdrawal", zap.Error(err))
			return ErrInternalError
		}
		err = tx.Commit()
		if err != nil {
			l.Error("Error commiting withdrawal", zap.Error(err))
			return ErrInternalError
		}
	} else {
		l.Info("Not enough funds")
		return ErrNotEnoughFunds
	}

	return nil
}

func (u *orderRepo) UpdateOrder(ctx context.Context, order models.Order) error {
	l := logr.FromContext(ctx)

	sqlStatement := `UPDATE orders SET status = $1, accrual = $2, processed_at = $3 WHERE order_id = ($4)`
	_, err := u.db.ExecContext(ctx, sqlStatement, order.Status, order.Accrual, order.ProcessedAt, order.OrderID)
	if err != nil {
		l.Error("Error updating order", zap.Error(err), zap.Any("order", order))
		return err
	}

	return nil
}

func (u *orderRepo) ListUnprocessedOrders(ctx context.Context, limit, offset int) ([]*models.Order, error) {
	l := logr.FromContext(ctx)

	sqlStatement := `SELECT order_id, status, tx_type, accrual, user_id, uploaded_at, processed_at
FROM orders 
WHERE tx_type = $1 AND status not in ($2, $3) LIMIT $4 OFFSET $5`

	rows, err := u.db.QueryContext(ctx, sqlStatement,
		models.DepositOrder,
		models.InvalidStatus,
		models.ProcessedStatus,
		limit,
		offset)

	if err != nil {
		if err == sql.ErrNoRows {
			return []*models.Order{}, nil
		}

		l.Error("Error querying for orders", zap.Error(err))
		return nil, ErrInternalError
	}
	err = rows.Err()
	if err != nil {
		l.Error("Error querying for orders", zap.Error(err))
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
			l.Error("Error scanning orders into object", zap.Error(err))
			return nil, ErrInternalError
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

func newOrderRepo(db *sql.DB, logger *zap.Logger) *orderRepo {
	if logger == nil {
		logger = logr.NewNoop()
	}
	return &orderRepo{db: db, log: logger}
}

func (u *orderRepo) CreateNewOrder(ctx context.Context, order models.Order) error {
	l := logr.FromContext(ctx)

	findOrderSQL := `SELECT user_id FROM orders WHERE order_id = $1`
	var userID int
	err := u.db.QueryRowContext(ctx, findOrderSQL, order.OrderID).Scan(&userID)
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
			l.Error("Error inserting order", zap.Error(err))
			return ErrInternalError
		}

		inserts, err := res.RowsAffected()
		if err != nil {
			l.Error("Error creating order", zap.Error(err))
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
		l.Error("Error creating order", zap.Error(err))
		return ErrInternalError
	}

	if userID == order.UserID {
		l.Info("Order already uploaded by current user")
		return ErrOrderAlreadyUploadedByCurrentUser
	} else {
		l.Info("Order already uploaded by another user")
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

func (u *orderRepo) CurrentBalance(ctx context.Context, userID int) (models.Balance, error) {
	var err error
	l := logr.FromContext(ctx)

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		l.Error("Could not begin tx", zap.Error(err))
	}
	defer tx.Rollback()

	sqlStatement := `SELECT COALESCE(SUM(accrual),0)
AS total FROM orders 
WHERE user_id = $1 AND tx_type = $2 AND status = $3`

	var deposit float64
	err = tx.QueryRowContext(ctx, sqlStatement, userID, models.DepositOrder, models.ProcessedStatus).Scan(&deposit)
	if err != nil {
		l.Error("Error quering deposit", zap.Error(err))
		return models.Balance{}, ErrInternalError
	}

	var withdrawal float64
	err = tx.QueryRowContext(ctx, sqlStatement, userID, models.WithdrawalOrder, models.ProcessedStatus).Scan(&withdrawal)
	if err != nil {
		l.Error("Error withdrawal deposit", zap.Error(err))
		return models.Balance{}, ErrInternalError
	}

	tx.Commit()
	return models.Balance{
		Current:   deposit - withdrawal,
		Withdrawn: withdrawal,
	}, nil
}
