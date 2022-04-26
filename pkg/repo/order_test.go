package repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/require"

	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
)

var testDb = "postgresql://practicum:practicum@localhost:5432"

func Test_orderRepo_CreateNewOrder(t *testing.T) {
	selectSql := `SELECT user_id FROM orders WHERE order_id = \$1`
	insertSql := `INSERT INTO orders \(order_id, status, tx_type, accrual, user_id, uploaded_at\)
		VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`

	order := models.Order{
		OrderID: 12345,
		Status:  "someStatus",
		TXType:  "someType",
		Accrual: 0,
		UserID:  8,
	}

	t.Run("insert single", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		s := mock.ExpectQuery(selectSql).WithArgs(order.OrderID)
		s.WillReturnError(sql.ErrNoRows)

		q := mock.ExpectExec(insertSql).
			WithArgs(
				order.OrderID,
				models.NewStatus,
				order.TXType,
				order.Accrual,
				order.UserID,
				sqlmock.AnyArg(),
			)
		q.WillReturnError(nil)
		q.WillReturnResult(sqlmock.NewResult(123, 1))

		repo, err := OrderRepo(db, newDevLogger(t))
		require.NoError(t, err)
		err = repo.CreateNewOrder(context.Background(), order)
		require.NoError(t, err)
	})

	t.Run("insert duplicate for current user", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		s := mock.ExpectQuery(selectSql).WithArgs(order.OrderID)
		s.WillReturnError(nil)
		s.WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(order.UserID))

		repo, err := OrderRepo(db, newDevLogger(t))
		require.NoError(t, err)
		err = repo.CreateNewOrder(context.Background(), order)
		require.ErrorIs(t, err, ErrOrderAlreadyUploadedByCurrentUser)
	})

	t.Run("insert already exists for another user", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		s := mock.ExpectQuery(selectSql).WithArgs(order.OrderID)
		s.WillReturnError(nil)
		s.WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1235))

		repo, err := OrderRepo(db, newDevLogger(t))
		require.NoError(t, err)
		err = repo.CreateNewOrder(context.Background(), order)
		require.ErrorIs(t, err, ErrOrderCreatedByAnotherUser)
	})
}

func Test_orderRepo_CurrentBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	log := newDevLogger(t)

	repo := orderRepo{db, log}

	uID := 3
	sum := 20
	q := mock.ExpectQuery(`SELECT COALESCE\(SUM\(accrual\),0\) AS total FROM orders WHERE user_id = \$1`).WithArgs(uID)
	q.WillReturnRows(mock.NewRows([]string{"total"}).AddRow(sum))
	q.WillReturnError(nil)

	balance, err := repo.CurrentBalance(context.Background(), uID)
	require.NoError(t, err)
	require.Equal(t, sum, balance)
}

func Test_orderRepo_queryOrders(t *testing.T) {
	columns := []string{
		"order_id",
		"status",
		"tx_type",
		"accrual",
		"user_id",
		"uploaded_at",
		"processed_at",
	}
	tests := []struct {
		name    string
		TXtype  models.OrderType
		wantErr bool
		err     error
		userID  int
		rows    *sqlmock.Rows
		orders  []*models.Order
	}{
		{
			"withdrawal",
			models.WithdrawalOrder,
			false,
			nil,
			2,
			sqlmock.NewRows(columns).AddRow(
				1,
				models.NewStatus,
				models.WithdrawalOrder,
				10,
				5,
				time.Date(1988, time.May, 10, 9, 0, 0, 0, time.UTC),
				time.Date(1988, time.May, 10, 9, 0, 0, 0, time.UTC),
			),
			[]*models.Order{{
				OrderID:     1,
				Status:      models.NewStatus,
				TXType:      models.WithdrawalOrder,
				Accrual:     10,
				UserID:      5,
				UploadedAt:  time.Date(1988, time.May, 10, 9, 0, 0, 0, time.UTC),
				ProcessedAt: time.Date(1988, time.May, 10, 9, 0, 0, 0, time.UTC),
			}},
		},
		{
			"deposit",
			models.DepositOrder,
			false,
			nil,
			2,
			sqlmock.NewRows(columns).AddRow(
				1,
				models.NewStatus,
				models.DepositOrder,
				10,
				5,
				time.Date(1988, time.May, 10, 9, 0, 0, 0, time.UTC),
				time.Date(1988, time.May, 10, 9, 0, 0, 0, time.UTC),
			),
			[]*models.Order{{
				OrderID:     1,
				Status:      models.NewStatus,
				TXType:      models.DepositOrder,
				Accrual:     10,
				UserID:      5,
				UploadedAt:  time.Date(1988, time.May, 10, 9, 0, 0, 0, time.UTC),
				ProcessedAt: time.Date(1988, time.May, 10, 9, 0, 0, 0, time.UTC),
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			log := newDevLogger(t)

			repo := orderRepo{db, log}

			sqlQuery := `SELECT order_id, status, tx_type, accrual, user_id, uploaded_at, processed_at
FROM orders 
WHERE user_id = \$1 AND tx_type = \$2`

			mock.ExpectQuery(sqlQuery).WithArgs(tt.userID, tt.TXtype).WillReturnRows(tt.rows)

			orders, err := repo.queryOrders(context.Background(), tt.userID, tt.TXtype)
			require.NoError(t, err)
			require.Equal(t, orders[0], tt.orders[0])

		})
	}
}
