package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/OmAsana/go-yapraktikum-final/pkg/logger"
)

var testDb = "postgresql://practicum:practicum@localhost:5432"

func newDevLogger(t *testing.T) *zap.Logger {
	t.Helper()
	devLogger, err := logger.NewDevLogger()
	if err != nil {
		t.Fatal(err)
	}
	return devLogger
}

func TestCreateUser(t *testing.T) {
	log := newDevLogger(t)

	type args struct {
		username, password string
	}

	tests := []struct {
		name    string
		wantErr bool
		err     Error
		args
	}{
		{
			"success",
			false,
			nil,
			args{
				"stepanar",
				"some_pass",
			},
		},
		{
			"duplicate",
			true,
			UserAlreadyExists,
			args{
				"stepanar",
				"some_pass",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			sqlQuery := `INSERT INTO users\(username, password_hash, created_at\) VALUES\(\$1, \$2, \$3\)`
			q := mock.ExpectPrepare(sqlQuery).
				ExpectExec().
				WithArgs(tt.args.username, tt.password, sqlmock.AnyArg())

			if tt.wantErr {
				q.WillReturnError(errors.New("SQLSTATE 23505"))
				q.WillReturnResult(sqlmock.NewResult(123, 123))
			} else {
				q.WillReturnError(nil)
				q.WillReturnResult(sqlmock.NewResult(123, 123))
			}
			userRepo := newUserRepo(db, log)
			err = userRepo.Create(context.TODO(), tt.args.username, tt.args.password)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestUserAuth(t *testing.T) {
	log := newDevLogger(t)

	type args struct {
		username, password string
	}
	tests := []struct {
		name    string
		wantErr bool
		userId  int
		err     Error
		args
	}{
		{
			"success",
			false,
			1,
			nil,
			args{
				"stepanar",
				"somepass",
			},
		},

		{
			"wrong pass",
			true,
			-1,
			UserAuthFailed,
			args{
				"stepanar",
				"somepass",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			sqlStatement := `SELECT user_id, password_hash FROM users WHERE username=\$1`
			q := mock.ExpectQuery(sqlStatement).
				WithArgs(tt.args.username)

			columns := []string{"user_id", "password_hash"}
			var rows *sqlmock.Rows
			if tt.wantErr {
				rows = mock.NewRows(columns).AddRow(tt.userId, tt.args.password+"lkjaslkfj")
			} else {
				rows = mock.NewRows(columns).AddRow(tt.userId, tt.args.password)
			}

			q.WillReturnRows(rows)
			userRepo := newUserRepo(db, log)
			id, err := userRepo.Authenticate(context.Background(), tt.args.username, tt.args.password)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.userId, id)

		})
	}
}
