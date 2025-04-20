package main

import (
	"sync/atomic"

	"github.com/ASRafalsky/telemetry/internal/storage"
)

type extendedRepository struct {
	*storage.MemStorage[string, []byte]
	newData atomic.Bool
}

func newExtendedRepository() *extendedRepository {
	return &extendedRepository{
		MemStorage: storage.New[string, []byte](),
	}
}

func (r *extendedRepository) Set(key string, value []byte) {
	r.MemStorage.Set(key, value)
	r.newData.Store(true)
}
