package repository

import (
	"context"
	"sync"

	"github.com/ASRafalsky/telemetry/internal/storage"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type RepositoryUnit struct {
	Mx sync.Mutex
	Repository
}

type Repository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Keys() []string
	Size() int
	Delete(k string)
}

func NewRepositories() map[string]*RepositoryUnit {
	return map[string]*RepositoryUnit{
		Gauge: {
			Repository: storage.New[string, []byte](),
		},
		Counter: {
			Repository: storage.New[string, []byte](),
		},
	}
}
