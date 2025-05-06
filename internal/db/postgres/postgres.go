package postgres

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sql.DB
}

func Open(param string) (DB, error) {
	db, err := sql.Open("pgx", param)
	return DB{DB: db}, err
}
func (db DB) Close() error {
	return db.DB.Close()
}

func (db DB) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}
