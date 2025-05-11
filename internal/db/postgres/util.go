package postgres

import (
	"context"
	"time"
)

func (d DB) WaitDBIsReady(ctx context.Context, tryCnt int, timeout time.Duration) error {
	var err error
	for range tryCnt {
		ctxPing, cancel := context.WithTimeout(ctx, timeout)
		err = d.Ping(ctxPing)
		cancel()
	}
	return err
}
