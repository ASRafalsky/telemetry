package repository

import (
	"context"

	"github.com/ASRafalsky/telemetry/internal/storage"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type Repository interface {
	Set(k string, v []byte)
	Get(k string) ([]byte, bool)
	ForEach(ctx context.Context, fn func(k string, v []byte) error) error
	Size() int
	Delete(k string)
}

func NewRepositories() map[string]Repository {
	return map[string]Repository{
		Gauge:   storage.New[string, []byte](),
		Counter: storage.New[string, []byte](),
	}
}
