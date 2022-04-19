package repo

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
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
	log.Info("blah")
	db, err := sql.Open("pgx", testDb)
	require.NoError(t, err)
	userRepo := newUserRepo(db, log)
	userRepo.Create(context.TODO(), "stepanar", "my_pass")
}
