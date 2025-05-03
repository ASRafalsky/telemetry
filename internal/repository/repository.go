package repository

import (
	"context"
	"sync/atomic"
)

type ExtendedRepository struct {
	dataStorage
	newData atomic.Bool
}

func NewExtendedRepository(storage dataStorage) *ExtendedRepository {
	return &ExtendedRepository{
		dataStorage: storage,
	}
}

func (r *ExtendedRepository) Set(key string, value []byte) {
	r.dataStorage.Set(key, value)
	r.newData.Store(true)
}

func (r *ExtendedRepository) IsReady() bool {
	return r.newData.Load()
}

func (r *ExtendedRepository) Ready() {
	r.newData.Store(true)
}

type dataStorage interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
	Delete(k string)
}
