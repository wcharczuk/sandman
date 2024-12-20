package db

import (
	"bytes"
	"sync"
)

/*
NewBufferPool returns a new Pool, which returns bytes buffers pre-sized to a given minimum size.

The purpose of a buffer pool is to reduce the number of gc collections incurred when using bytes buffers
repeatedly; instead of marking the buffer as to be collected, it is returned to the pool to be re-used.

Example:

	pool := bufferutil.NewBufferPool(1024) // pre-allocate 1024 bytes per buffer.

	func() {
		buf := pool.Get()
		defer pool.Put(buf)

		// do things with the buffer ...
	}()
*/
func NewBufferPool(defaultBufferSize int) *BufferPool {
	return &BufferPool{
		Pool: sync.Pool{New: func() interface{} {
			b := bytes.NewBuffer(make([]byte, defaultBufferSize))
			b.Reset()
			return b
		}},
	}
}

// BufferPool is a sync.Pool of bytes.Buffer.
type BufferPool struct {
	sync.Pool
}

// Get returns a pooled bytes.Buffer instance.
func (p *BufferPool) Get() *bytes.Buffer {
	return p.Pool.Get().(*bytes.Buffer)
}

// Put returns the pooled instance.
func (p *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	p.Pool.Put(b)
}
