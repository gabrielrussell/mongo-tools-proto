package util
// TODO this should probably be its own package,
// to reduce binary size for some tools

import (
	"fmt"
	"sync"
)

// BufferPool is a construct for generating and reusing buffers of a
// given size. Useful for avoiding generating too many temporary
// buffers during runtime, which can anger the garbage collector.
type BufferPool struct {
	size int
	p    *sync.Pool //REQUIRES >= go1.3
}

// returns a "New" function for use by sync.Pool
func newByteBufferFunc(size int) func() interface{} {
	return func() interface{} {
		return make([]byte, size)
	}
}

// NewBufferPool returns an initialized BufferPool for
// buffers of the supplied number of bytes.
func NewBufferPool(bytes int) *BufferPool {
	if bytes < 0 {
		panic("cannot create BufferPool of negative size")
	}
	bp := &BufferPool{
		size: bytes,
		p: &sync.Pool{
			New: newByteBufferFunc(bytes),
		},
	}
	return bp
}

// Get returns a new or recycled buffer form the pool.
func (bp *BufferPool) Get() []byte {
	return bp.p.Get().([]byte)
}

// Put returns the supplied slice back to the buffer.
// Panics if the buffer is of improper size.
func (bp *BufferPool) Put(buffer []byte) {
	if len(buffer) != bp.size {
		panic(fmt.Sprintf(
			"attempting to return a byte buffer of size %v to a BufferPool of size %v",
			len(buffer), bp.size))
	}
	bp.p.Put(buffer)
}
