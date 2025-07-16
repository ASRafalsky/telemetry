package pool

import (
	"bytes"
	"sync"
)

type LimitedPool struct {
	pool        sync.Pool
	maxSize     int
	defaultSize int
}

func NewLimitedPool(maxSize int, defaultSize int) *LimitedPool {
	return &LimitedPool{
		maxSize:     maxSize,
		defaultSize: defaultSize,
	}
}

func (p *LimitedPool) Get() *bytes.Buffer {
	buf := p.pool.Get()
	if buf == nil {
		return bytes.NewBuffer(make([]byte, 0, p.defaultSize))
	}
	buf.(*bytes.Buffer).Reset()
	return buf.(*bytes.Buffer)
}

func (p *LimitedPool) Put(buf *bytes.Buffer) {
	if buf.Len() > p.maxSize {
		return
	}
	p.pool.Put(buf)
}
