// Package pool includes configurable sync pools for various purposes.
package pool

import (
	"bytes"
	"sync"
)

// LimitedPool describes bytes.Buffer sync Pool with max buffer size limit.
// This limit is useful to avoid storing too large buffers in the pool.
type LimitedPool struct {
	pool    sync.Pool // You will be surprised, but this is sync Pool.
	maxSize int       // Max size limit in bytes.
}

// NewLimitedPool returns the LimitedPool instance with max size limit in bytes.
func NewLimitedPool(maxSize int) *LimitedPool {
	return &LimitedPool{
		maxSize: maxSize,
	}
}

// Get returns bytes.Buffer instance from LimitedPool.
// If the pool is empty, a new one is created, otherwise the flushed buffer is returned from the pool.
func (p *LimitedPool) Get() *bytes.Buffer {
	buf := p.pool.Get()
	if buf == nil {
		return bytes.NewBuffer(nil)
	}
	buf.(*bytes.Buffer).Reset()
	return buf.(*bytes.Buffer)
}

// Put puts the buffer into the pool if its size is less than max size limit.
func (p *LimitedPool) Put(buf *bytes.Buffer) {
	if buf.Len() > p.maxSize {
		return
	}
	p.pool.Put(buf)
}
