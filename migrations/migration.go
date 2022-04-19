package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var embedMigrations embed.FS

func ApplyMigrations(uri string) error {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "sql"); err != nil {
		return err
	}
	return nil
}
