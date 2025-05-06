package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
	ts := time.Now()
	fmt.Println("ololo", ts)
	defer fmt.Println("ololo", time.Since(ts))
	return db.DB.PingContext(ctx)
}
