package repo

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	logr "github.com/OmAsana/go-yapraktikum-final/pkg/logger"
)

var _ UserRepository = (*userRepo)(nil)

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

func (u *userRepo) Create(ctx context.Context, username string, password string) (int, error) {
	var err error
	l := logr.FromContext(ctx)
	defer func() {
		if err != nil {
			l.Error("could not create user", zap.Error(err))
		}
	}()

	l.Info("creating user")

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return -1, ErrInternalError
	}

	now := time.Now()
	prepareContext, err := u.db.PrepareContext(ctx, "INSERT INTO users(username, password_hash, created_at) VALUES($1, $2, $3) RETURNING user_id")
	if err != nil {
		return -1, ErrInternalError
	}

	var id int
	err = prepareContext.QueryRowContext(ctx, username, string(hash), now).Scan(&id)
	if err != nil {
		// duplicate key error
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			return -1, ErrUserAlreadyExists
		}

		return -1, ErrInternalError
	}
	return id, nil
}

func (u *userRepo) Authenticate(ctx context.Context, username string, password string) (int, error) {
	var err error
	l := logr.FromContext(ctx)
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
		return -1, ErrUserNotFound
	case err != nil:
		return -1, ErrInternalError
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return -1, ErrUserAuthFailed
	}

	return id, nil
}
