package mint

import (
	"bytes"
	"sync"
)

//BufferPool #
type BufferPool struct {
	p *sync.Pool
}

//NewBufferPool #
func NewBufferPool() *BufferPool {
	return &BufferPool{
		p: &sync.Pool{
			New: func() interface{} { return new(bytes.Buffer) },
		},
	}
}

//Get #
func (bpool *BufferPool) Get() *bytes.Buffer {
	return bpool.p.Get().(*bytes.Buffer)
}

//Put #
func (bpool *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	bpool.p.Put(b)
}
