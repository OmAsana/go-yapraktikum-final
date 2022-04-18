package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/OmAsana/go-yapraktikum-final/migrations"
)

var (
	flags       = flag.NewFlagSet("migrate", flag.ExitOnError)
	databaseDSN = flags.String("d", "", "Postgre database connection string (required)")
)

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("migration: failed to parse flags: %v\n", err)
	}
	if *databaseDSN == "" {
		flags.Usage()
		os.Exit(1)
	}

	db, err := sql.Open("pgx", *databaseDSN)
	if err != nil {
		log.Fatalf("migration: failed to open DB: %v\n", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("migration: failed to close DB: %v\n", err)
		}
	}()

	if err := migrations.ApplyMigrations(db); err != nil {
		log.Fatalf("migration: failed to apply migration: %v\n", err)
	}

}
