package repo

import (
	"database/sql"

	"go.uber.org/zap"
)

func UserRepo(db *sql.DB, log *zap.Logger) (User, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return newUserRepo(db, log), nil
}

func OrderRepo(db *sql.DB, log *zap.Logger) (Order, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return newOrderRepo(db, log), nil
}
