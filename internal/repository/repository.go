// Package repository contains repository with database and cache extension, and all necessary methods for interaction
// with it.
package repository

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ASRafalsky/telemetry/internal/db/postgres"
	"github.com/ASRafalsky/telemetry/internal/types"
)

// ExtendedRepository describes ExtendedRepository.
type ExtendedRepository struct {
	db      db
	cache   cache
	newData atomic.Bool
}

// NewExtendedRepository creates new ExtendedRepository instance with cache.
func NewExtendedRepository(cache cache) *ExtendedRepository {
	return &ExtendedRepository{
		cache: cache,
	}
}

// UseDB adds db to the ExtendedRepository instance.
func (r *ExtendedRepository) UseDB(db db) {
	r.db = db
}

// Set sets key value pair.
func (r *ExtendedRepository) Set(key string, value []byte) {
	r.cache.Set(key, value)
	r.newData.Store(true)
}

// IsReady returns true if ExtendedRepository contains new data.
func (r *ExtendedRepository) IsReady() bool {
	return r.newData.Load()
}

// Ready set ExtendedRepository to the no new data state.
func (r *ExtendedRepository) Ready() {
	r.newData.Store(false)
}

// Ping checks db if it sets.
func (r *ExtendedRepository) Ping(ctx context.Context) error {
	if r.db == nil {
		return nil
	}
	return r.db.Ping(ctx)
}

// Get returns data from repository and error if anything went wrong. If this entry doesn't exist it returns
// nil, nil (I know, this is bad behavior, don't follow me).
//
// TODO(ASRafalsky): Do something about return nil, nil.
func (r *ExtendedRepository) Get(ctx context.Context, key string) ([]byte, error) {
	if val, ok := r.cache.Get(key); ok {
		r.Set(key, val)
		return val, nil
	}
	if r.db != nil {
		var (
			val []byte
			err error
		)
		err = withRetryOnErr(ctx, 3, func() error {
			val, err = r.db.Get(ctx, key)
			return err
		})
		return val, err
	}
	return nil, nil
}

// Delete removes entry from repository by key and returns error if anything went wrong.
func (r *ExtendedRepository) Delete(ctx context.Context, key string) error {
	r.cache.Delete(key)
	if r.db != nil {
		return withRetryOnErr(ctx, 3, func() error {
			return r.db.Delete(ctx, key)
		})
	}
	return nil
}

// ForEach calls fn for each entry in the repository. And you know what? Yeah! It returns error if anything went wrong.
func (r *ExtendedRepository) ForEach(ctx context.Context, fn func(k string, v []byte) error) error {
	if r.db != nil {
		if r.newData.Load() {
			if err := r.Sync(ctx); err != nil {
				return err
			}
		}
		return withRetryOnErr(ctx, 3, func() error { return r.db.ForEach(ctx, fn) })
	}
	return r.cache.ForEach(ctx, fn)
}

// Size returns the number of entries in the repository. If it uses db it will be number entries from the db,
// else from the cache.
func (r *ExtendedRepository) Size() (int, error) {
	if r.db != nil {
		var (
			sz  int
			err error
		)
		err = withRetryOnErr(context.Background(), 3, func() error {
			sz, err = r.db.Size()
			return err
		})
		return sz, err
	}
	return r.cache.Size(), nil
}

// Sync syncs cache with db.
func (r *ExtendedRepository) Sync(ctx context.Context) error {
	if r.db == nil {
		return nil
	}
	p := make([]postgres.Pair, r.cache.Size())
	if err := r.cache.DropFn(ctx, func(k string, v []byte) (bool, error) {
		var drop bool
		switch {
		case strings.HasPrefix(k, types.GaugeType):
			drop = true
		case strings.HasPrefix(k, types.CounterType): // Do not drop it because we need this value.
		default:
			return false, nil
		}
		p = append(p, postgres.Pair{ID: k, Payload: v})
		return drop, nil
	}); err != nil {
		return err
	}
	r.Ready()
	return withRetryOnErr(ctx, 3, func() error { return r.db.SetBatch(ctx, p) })
}

// CacheSize returns the number of entries in the cache.
func (r *ExtendedRepository) CacheSize() int {
	return r.cache.Size()
}

type cache interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	DropFn(ctx context.Context, fn func(k string, v []byte) (bool, error)) error
	Size() int
	Delete(k string)
}

type db interface {
	Ping(ctx context.Context) error
	Set(ctx context.Context, key string, data []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	SetBatch(ctx context.Context, batch []postgres.Pair) error
	Close() error
	Size() (int, error)
}

func withRetryOnErr(ctx context.Context, cnt int, fn func() error) error {
	err := fn()
	if err != nil && (errors.Is(err, postgres.ErrConnectionIssue) || errors.Is(err, postgres.ErrTransaction)) {
		cnt--
		wait := 1
		ticker := time.NewTicker(time.Duration(wait) * time.Second)
		defer ticker.Stop()
		for cnt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				err = fn()
				if err == nil {
					return nil
				}
				wait += 2
				ticker.Reset(time.Duration(wait) * time.Second)
				cnt--
			}
		}
	}
	return err
}
