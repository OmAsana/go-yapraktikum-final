package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/OmAsana/go-yapraktikum-final/pkg/logger"
)

//var testDb = "postgresql://practicum:practicum@localhost:5432"

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
			ErrUserAlreadyExists,
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
			sqlQuery := `INSERT INTO users\(username, password_hash, created_at\) VALUES\(\$1, \$2, \$3\) RETURNING user_id`
			q := mock.ExpectPrepare(sqlQuery).
				ExpectQuery().
				WithArgs(tt.args.username, sqlmock.AnyArg(), sqlmock.AnyArg())

			if tt.wantErr {
				q.WillReturnError(errors.New("SQLSTATE 23505"))
				//q.WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow())
			} else {
				q.WillReturnError(nil)
				q.WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))
			}
			userRepo := newUserRepo(db, log)
			id, err := userRepo.Create(context.TODO(), tt.args.username, tt.args.password)
			assert.ErrorIs(t, err, tt.err)
			if tt.wantErr {
				require.Equal(t, id, -1)
			} else {
				require.Equal(t, id, 1)
			}
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
		userID  int
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
			ErrUserAuthFailed,
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
				rows = mock.NewRows(columns).AddRow(tt.userID, helpGenerateHash(t, tt.args.password+"some_random_str"))
			} else {
				rows = mock.NewRows(columns).AddRow(tt.userID, helpGenerateHash(t, tt.args.password))
			}

			q.WillReturnRows(rows)
			userRepo := newUserRepo(db, log)
			id, err := userRepo.Authenticate(context.Background(), tt.args.username, tt.args.password)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.userID, id)

		})
	}
}

func helpGenerateHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	require.NoError(t, err)
	return string(hash)
}
