package syncio

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/travelgateX/go-io/syncio/internal"
)

var _ io.WriteCloser = &Buffer{}

// Buffer is a buffer which implements io.WriteCloser methods where writes can be done
// concurrently, the writes store the data in a buffer thats its later flushed to an underlying writer
// when its full or ticks.
// The underlying writer might receive concurrent writes.
// Closing blocks the caller until all writes finish.
type Buffer struct {
	bufmu sync.Mutex
	buf   *internal.Buffer
	// buffer write operations to the underlaying writer are potentially slow, a new buffer
	// takes the stage when the current is sent to write
	pool *internal.BufferPool

	writer        io.Writer
	bufSize       int
	poolSize      int
	flushInterval time.Duration

	// control flag to not flush per tick if a flush is
	// already done by full buffer
	flushedBetweenTicks bool

	closed bool
	// wg is a group of writes in progress
	wg sync.WaitGroup

	stats Stats
}

// NewBuffer wraps a writer with a buffer layer that will write to an underlying writer
// when its buffer is full or a timer ticks
// Call Close to free goroutines, Close blocks until all buffers flush, calling Close and then Write won't panic
func NewBuffer(w io.Writer, options ...BufferOption) *Buffer {
	const (
		defaultBufSize  = 4096
		defaultPoolSize = 2
	)

	tb := &Buffer{writer: w}
	for _, o := range options {
		o(tb)
	}

	if tb.bufSize == 0 {
		tb.bufSize = defaultBufSize
	}
	if tb.poolSize == 0 {
		tb.poolSize = defaultPoolSize
	}
	tb.pool = internal.NewBufferPool(tb.poolSize, tb.bufSize)
	tb.buf = tb.getBuffer()

	if tb.flushInterval > 0 {
		go func() {
			t := time.NewTicker(tb.flushInterval)
			for !tb.closed {
				select {
				case <-t.C:
					tb.bufmu.Lock()
					if !tb.flushedBetweenTicks {
						tb.flush()
					} else {
						tb.flushedBetweenTicks = false
					}
					tb.bufmu.Unlock()
				}
			}
			t.Stop()
		}()
	}

	return tb
}

// BufferOption are optional configurations used on a Buffer instantiation
type BufferOption func(*Buffer)

// SetBufferSize to set the maximum size of a buffer, it flushes when its full
func SetBufferSize(s int) BufferOption {
	return func(b *Buffer) {
		b.bufSize = s
	}
}

// SetBufferPoolSize to set the minimum number of buffers
// that will be retained in the pool which won't be garbage collected
func SetBufferPoolSize(s int) BufferOption {
	return func(b *Buffer) {
		b.poolSize = s
	}
}

// SetFlushInterval to set the interval between flushes, if this options is used,
// the buffers will flush when full or when a ticks happen
func SetFlushInterval(d time.Duration) BufferOption {
	return func(b *Buffer) {
		b.flushInterval = d
	}
}

// ErrWriteOnClosed is returned when a write is done after closing
var ErrWriteOnClosed = errors.New("write on closed writer")

// Write enqueues the data to be buffered
func (tb *Buffer) Write(p []byte) (int, error) {
	lenP := len(p)

	tb.bufmu.Lock()
	defer tb.bufmu.Unlock()
	if tb.closed {
		return 0, ErrWriteOnClosed
	}

	// case when p is bigger than the buffer size:
	// copy to intermediate buffer to make sure that the write is not blocked by the underlying write
	// and p is not retained, this is an unexpected use case, TickedBuffer should
	// have buffers with size multiple times higher than a single write...
	// TODO: improve the performance of this usecase ???
	if lenP >= tb.bufSize {
		b := make([]byte, lenP)
		copy(b, p)
		tb.wg.Add(1)
		go func() {
			_, err := tb.writer.Write(b)
			if err != nil {
				atomic.AddInt32(&tb.stats.FlushErrors, 1)
			}
			tb.wg.Done()
		}()
		return lenP, nil
	}

	if lenP > tb.buf.Available() {
		tb.flush()
		tb.flushedBetweenTicks = true
	}
	tb.buf.Write(p)

	return lenP, nil
}

// Close is concurrent safe and blocks until the remaining data in buffer is flushed
func (tb *Buffer) Close() error {
	tb.bufmu.Lock()
	tb.closed = true
	tb.bufmu.Unlock()

	// flush remaining data in buffers
	tb.flush()

	tb.wg.Wait()
	return nil
}

// flush writes all the data of the current buffer to the underlying writer
// the buffer used to write is put in background and its sent back to the
// buffer pool when its operation finishes. A new buffer is obtained to continue
// serving incoming writes.
func (tb *Buffer) flush() {
	if tb.buf.Buffered() == 0 {
		return
	}
	buf := tb.buf
	tb.buf = tb.getBuffer()
	tb.wg.Add(1)
	go func() {
		_, err := buf.WriteTo(tb.writer)
		if err != nil {
			atomic.AddInt32(&tb.stats.FlushErrors, 1)
		}
		tb.pool.Put(buf)
		tb.wg.Done()
	}()
}

func (tb *Buffer) getBuffer() *internal.Buffer {
	b, alloc := tb.pool.Get()
	if alloc {
		tb.stats.BufferAllocs++
	}
	return b
}

// Stats contains performance statistics, some of the settings for this writer
// can be hard to profile and depend on the environment and use case, these
// stats are meant to help to adjust the settings for a better performance
type Stats struct {
	// BufferAllocs is the number of new buffers
	// that had to be created because the pool had not enough
	// free buffers. Note that a pool is initialized without buffers.
	BufferAllocs int32
	// Count of errors obtained trying to write to the underlying writer
	FlushErrors int32
}

// Stats returns a copy of the current writer stats
func (tb *Buffer) Stats() Stats {
	return tb.stats
}
