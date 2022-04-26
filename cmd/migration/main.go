package main

import (
	"flag"
	"log"
	"os"

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

	if err := migrations.ApplyMigrations(*databaseDSN); err != nil {
		log.Fatalf("migration: failed to apply migration: %v\n", err)
	}

}
