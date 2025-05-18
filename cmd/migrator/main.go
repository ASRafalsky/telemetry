package main

import (
	"errors"
	"flag"
	"fmt"
	// migration lib
	"github.com/golang-migrate/migrate/v4"
	// driver for migration applying postgres
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// driver to get migrations from files (*.sql in our case)
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	var migrationsPath, migrationsTable, databaseDSN string

	flag.StringVar(
		&databaseDSN,
		"d",
		"",
		"database-dsn",
	)

	flag.StringVar(
		&migrationsPath,
		"p",
		"",
		"path to migrations",
	)
	flag.StringVar(
		&migrationsTable,
		"t",
		"migrations",
		"name of migration table, where migrator writes own data",
	)
	flag.Parse()

	if migrationsPath == "" {
		panic("migrations path is required")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseDSN,
	)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}
	fmt.Println("migrations applied successfully")
}
