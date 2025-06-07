package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/multierr"
)

type DB struct {
	*sql.DB
	m migrate.Migrate
}

var (
	ErrConnectionIssue = errors.New("database connection issue")
	ErrTransaction     = errors.New("database transaction issue")
)

func Open(param string) (DB, error) {
	db, err := sql.Open("pgx", param)
	return DB{DB: db}, err
}
func (d DB) Close() error {
	return d.DB.Close()
}

func (d DB) Ping(ctx context.Context) error {
	return d.DB.PingContext(ctx)
}

// Bootstrap prepares DB.
func (d DB) Bootstrap(ctx context.Context) error {
	return d.bootstrap(ctx, `CREATE TABLE IF NOT EXISTS metrics (
            id VARCHAR(128) PRIMARY KEY,
            payload JSONB);`)
}

func (d DB) bootstrap(ctx context.Context, query string) (err error) {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	defer func() {
		if errRollback := tx.Rollback(); err != nil && errRollback != nil {
			err = multierr.Append(err, errRollback)
		}
	}()

	// Create metrics table.
	if _, errCreate := tx.ExecContext(ctx, query); err != nil {
		err = multierr.Append(err, errCreate)
	}

	if errCommit := tx.Commit(); errCommit != nil {
		err = multierr.Append(err, errCommit)
	}
	return
}

func (d DB) Set(ctx context.Context, key string, data []byte) error {
	return d.set(ctx,
		`INSERT INTO metrics (id, payload) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET payload = $2`, key, data)
}

func (d DB) set(ctx context.Context, query, key string, data []byte) error {
	_, err := d.ExecContext(ctx, query, key, data)
	return errorHandler(err)
}

func (d DB) Get(ctx context.Context, key string) ([]byte, error) {
	return d.get(ctx, `SELECT payload FROM metrics WHERE id = $1`, key)
}

func (d DB) get(ctx context.Context, query, key string) ([]byte, error) {
	var payload []byte
	err := d.QueryRowContext(ctx, query, key).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, errorHandler(err)
	}
	return payload, nil
}

func (d DB) Delete(ctx context.Context, key string) error {
	return d.deleteEntrie(ctx, `DELETE FROM metrics WHERE id = $1`, key)
}

func (d DB) deleteEntrie(ctx context.Context, query, key string) error {
	_, err := d.ExecContext(ctx, query, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return errorHandler(err)
}

func (d DB) ForEach(ctx context.Context, fn func(k string, v []byte) error) error {
	return errorHandler(d.forEach(ctx, `SELECT id, payload FROM metrics ORDER BY id LIMIT $1 OFFSET $2`, fn))
}

const BatchSz = 1000

func (d DB) forEach(ctx context.Context, query string, fn func(k string, v []byte) error) (err error) {
	offset := 0
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if errRollback := tx.Rollback(); err != nil && errRollback != nil {
			err = multierr.Append(err, errRollback)
		}
	}()

	for {
		rows, err := tx.QueryContext(ctx, query, BatchSz, offset)
		if err != nil {
			return err
		}

		var processed int
		for rows.Next() {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var id string
			var payload []byte
			if err = rows.Scan(&id, &payload); err != nil {
				if closeErr := rows.Close(); closeErr != nil {
					err = multierr.Append(err, closeErr)
				}
				return err
			}
			if err = fn(id, payload); err != nil {
				if closeErr := rows.Close(); closeErr != nil {
					err = multierr.Append(err, closeErr)
				}
				return err
			}
			processed++
		}
		if err = rows.Err(); err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				err = multierr.Append(err, closeErr)
			}
			return err
		}
		err = rows.Close()
		if processed < BatchSz {
			if errCommit := tx.Commit(); err != nil {
				err = multierr.Append(err, errCommit)
			}
			return err
		}
		offset += BatchSz
	}
}

func (d DB) Size() (int, error) {
	var sz int
	err := d.QueryRow(`SELECT COUNT(*) FROM metrics`).Scan(&sz)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, errorHandler(err)
	}
	return sz, nil
}

type Pair struct {
	ID      string
	Payload []byte
}

func (d DB) SetBatch(ctx context.Context, batch []Pair) error {
	return errorHandler(d.setBatch(ctx,
		`INSERT INTO metrics (id, payload) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET payload = $2`, batch))
}

func (d DB) setBatch(ctx context.Context, query string, batch []Pair) (err error) {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if errRollback := tx.Rollback(); err != nil && errRollback != nil {
			err = multierr.Append(err, errRollback)
		}
	}()

	for i := range batch {
		_, err = tx.ExecContext(ctx, query, batch[i].ID, batch[i].Payload)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	return
}

func errorHandler(err error) error {
	var pgErr *pgconn.PgError
	if err != nil && errors.As(err, &pgErr) {
		switch {
		case pgerrcode.IsInvalidTransactionInitiation(pgErr.Code),
			pgerrcode.IsInvalidTransactionState(pgErr.Code),
			pgerrcode.IsInvalidTransactionTermination(pgErr.Code):
			return ErrTransaction
		case pgerrcode.IsConnectionException(pgErr.Code):
			return ErrConnectionIssue

		}
	}
	return err
}
