package internal

// BufferPool is a pool of buffers implementation where the buffers
// can't be garbage collected. https://golang.org/doc/effective_go.html#leaky_buffer
type BufferPool struct {
	free   chan *Buffer
	bufcap int
}

// NewBufferPool instances a bufferPool with 'size' buffers,
// buffers will be allocated with a 'bufcap' capacity
func NewBufferPool(size, bufcap int) *BufferPool {
	return &BufferPool{
		free:   make(chan *Buffer, size),
		bufcap: bufcap,
	}
}

// Get returns an available buffer, if any, a new one will be allocated.
// Returns a bool indicating if an allocation happened
func (p *BufferPool) Get() (*Buffer, bool) {
	select {
	case buf := <-p.free:
		// got one
		return buf, false
	default:
		// there aren't free buffers, allocate new one
		return newBuffer(p.bufcap), true
	}
}

// Put returns a buffer, its dropped on the floor if the pool is full
func (p *BufferPool) Put(b *Buffer) {
	b.Reset()
	select {
	case p.free <- b:
		// reuse buffer
	default:
		// free list full; drop
	}
}
