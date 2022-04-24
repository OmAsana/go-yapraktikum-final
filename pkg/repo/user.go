package repo

import (
	"context"
	"database/sql"
	"strings"
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
		return ErrInternalError
	}
	_, err = prepareContext.ExecContext(ctx, username, pwdHash, now)
	if err != nil {
		// duplicate key error
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			return ErrUserAlreadyExists
		}

		return ErrInternalError
	}
	return nil
}

func (u *userRepo) Authenticate(ctx context.Context, username string, pwdHash string) (int, Error) {
	var err error
	l := u.log.With(zap.String("username", username))
	defer func() {
		if err != nil {
			l.Error("could not authenticate user", zap.Error(err))
		}
	}()

	sqlStatement := `SELECT user_id, password_hash FROM users WHERE username=$1`

	var hash string
	var id int
	err = u.db.QueryRowContext(ctx, sqlStatement, username).Scan(&id, &hash)
	switch {
	case err == sql.ErrNoRows:
		u.log.Error("user does not exist")
		return -1, ErrUserAuthFailed
	case err != nil:
		return -1, ErrInternalError
	}

	if hash != pwdHash {
		u.log.Error("wrong password")
		return -1, ErrUserAuthFailed
	}

	return id, nil
}
