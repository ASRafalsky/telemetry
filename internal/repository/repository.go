package repository

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/ASRafalsky/telemetry/internal/db/postgres"
	"github.com/ASRafalsky/telemetry/internal/handlers"
)

type ExtendedRepository struct {
	db      db
	cache   cache
	newData atomic.Bool
}

func NewExtendedRepository(cache cache) *ExtendedRepository {
	return &ExtendedRepository{
		cache: cache,
	}
}

func (r *ExtendedRepository) UseDB(db db) {
	r.db = db
}

func (r *ExtendedRepository) Set(key string, value []byte) {
	r.cache.Set(key, value)
	r.newData.Store(true)
}

func (r *ExtendedRepository) IsReady() bool {
	return r.newData.Load()
}

func (r *ExtendedRepository) Ready() {
	r.newData.Store(false)
}

func (r *ExtendedRepository) Ping(ctx context.Context) error {
	if r.db == nil {
		return nil
	}
	return r.db.Ping(ctx)
}

func (r *ExtendedRepository) Get(ctx context.Context, key string) ([]byte, error) {
	if val, ok := r.cache.Get(key); ok {
		r.Set(key, val)
		return val, nil
	}
	if r.db != nil {
		return r.db.Get(ctx, key)
	}
	return nil, nil
}

func (r *ExtendedRepository) Delete(ctx context.Context, key string) error {
	r.cache.Delete(key)
	if r.db != nil {
		return r.db.Delete(ctx, key)
	}
	return nil
}

func (r *ExtendedRepository) ForEach(ctx context.Context, fn func(k string, v []byte) error) error {
	if r.db != nil {
		if r.newData.Load() {
			if err := r.Sync(ctx); err != nil {
				return err
			}
		}
		return r.db.ForEach(ctx, fn)
	}
	return r.cache.ForEach(ctx, fn)
}

func (r *ExtendedRepository) Size() (int, error) {
	if r.db != nil {
		return r.db.Size()
	}
	return r.cache.Size(), nil
}

func (r *ExtendedRepository) Sync(ctx context.Context) error {
	if r.db == nil {
		return nil
	}
	p := make([]postgres.Pair, r.cache.Size())
	if err := r.cache.DropFn(ctx, func(k string, v []byte) (bool, error) {
		var drop bool
		switch {
		case strings.HasPrefix(k, handlers.Gauge):
			drop = true
		case strings.HasPrefix(k, handlers.Counter): // Do not drop it because we need this value.
		default:
			return false, nil
		}
		p = append(p, postgres.Pair{ID: k, Payload: v})
		return drop, nil
	}); err != nil {
		return err
	}
	r.Ready()
	return r.db.SetBatch(ctx, p)
}

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
