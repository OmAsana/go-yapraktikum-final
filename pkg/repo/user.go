package repo

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"

	logr "github.com/OmAsana/go-yapraktikum-final/pkg/logger"
)

var _ User = (*userRepo)(nil)

type userRepo struct {
	db  *sql.DB
	log *zap.Logger
}

func newUserRepo(db *sql.DB, logger *zap.Logger) *userRepo {
	if logger == nil {
		logger = logr.NewNoop()
	}
	return &userRepo{db: db, log: logger}
}

func (u *userRepo) Create(ctx context.Context, username string, pwdHash string) Error {
	var err error
	l := u.log.With(zap.String("username", username))
	defer func() {
		if err != nil {
			l.Error("could not create user", zap.Error(err))
		}
	}()

	l.Info("creating user")
	now := time.Now()
	prepareContext, err := u.db.PrepareContext(ctx, "INSERT INTO users(username, password_hash, created_at) VALUES($1, $2, $3)")
	if err != nil {
		return InternalError
	}
	_, err = prepareContext.ExecContext(ctx, username, pwdHash, now)
	if err != nil {
		return InternalError
	}
	return nil
}

func (u *userRepo) Authenticate(ctx context.Context, username string, pwdHash string) Error {
	//TODO implement me
	panic("implement me")
}

func (u *userRepo) CurrentBalance(ctx context.Context, username string) int {
	//TODO implement me
	panic("implement me")
}
