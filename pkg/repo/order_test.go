package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/require"

	"github.com/OmAsana/go-yapraktikum-final/pkg/models"
)

//var testDb = "postgresql://practicum:practicum@localhost:5432"

func Test_orderRepo_CreateNewOrder(t *testing.T) {

	tests := []struct {
		name    string
		wantErr bool
		err     Error
		order   models.Order
	}{
		{
			"success",
			false,
			nil,
			models.Order{
				OrderId: 123,
				Status:  "someStatus",
				TXType:  "someType",
				Accrual: 0,
				UserId:  1,
			},
		},
		{
			"dup order",
			true,
			DuplicateOrder,
			models.Order{
				OrderId: 123,
				Status:  "NEW",
				TXType:  "someType",
				Accrual: 0,
				UserId:  1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			q := mock.ExpectExec(`INSERT INTO orders (.+) VALUES (.+)`).
				WithArgs(
					tt.order.OrderId,
					models.NewStatus,
					tt.order.TXType,
					tt.order.Accrual,
					tt.order.UserId,
					sqlmock.AnyArg(),
				)

			if tt.wantErr {
				q.WillReturnError(errors.New("SQLSTATE 23505"))
				q.WillReturnResult(sqlmock.NewResult(123, 0))
			} else {
				q.WillReturnError(nil)
				q.WillReturnResult(sqlmock.NewResult(123, 1))
			}

			log := newDevLogger(t)
			repo := orderRepo{db, log}
			err = repo.CreateNewOrder(context.Background(), tt.order)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}

func Test_orderRepo_CurrentBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	log := newDevLogger(t)

	repo := orderRepo{db, log}

	uId := 3
	sum := 20
	q := mock.ExpectQuery("^SELECT COALESCE\\(SUM\\(accrual\\),0\\) AS total from orders where user_id = \\$1").WithArgs(uId)
	q.WillReturnRows(mock.NewRows([]string{"total"}).AddRow(sum))
	q.WillReturnError(nil)

	balance, err := repo.CurrentBalance(context.Background(), uId)
	require.NoError(t, err)
	require.Equal(t, sum, balance)
}

//func Test_orderRepo_queryOrders(t *testing.T) {
//	columns := []string{
//		"order_id",
//		"status",
//		"tx_type",
//		"accrual",
//		"user_id",
//		"uploaded_at",
//		"processed_at",
//	}
//	tests := []struct {
//		name    string
//		wantErr bool
//		err     Error
//		userId  int
//		rows    []*sqlmock.Rows
//		orders  []*models.Order
//	}{
//		{
//			"depost",
//			false,
//			nil,
//			2,
//			sqlmock.NewRows(columns).AddRow(),
//			[]*models.Order{},
//		},
//	}
//	//db, err := sql.Open("pgx", testDb)
//	//require.NoError(t, err)
//	//
//	//log := newDevLogger(t)
//	//repo := orderRepo{db, log}
//	//
//	//orders, err := repo.queryOrders(context.Background(), 2, models.DepositOrder)
//	//require.NoError(t, err)
//	//
//	//for _, o := range orders {
//	//	fmt.Println(o)
//	//}
//}
