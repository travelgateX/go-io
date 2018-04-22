// Package asyncio provides asynchronous non-bloking io operations
package asyncio

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/travelgateX/go-io/asyncio/internal"
)

var _ io.WriteCloser = &TickedBuffer{}

// TickedBuffer is a buffer which implements io.WriteCloser methods where writes can be done
// concurrently, the writes store the data in a buffer thats its later flushed to an underlying writer
// when its full or ticks.
// The underlying writer might receive concurrent writes.
// Closing blocks the caller until all writes finish.
type TickedBuffer struct {
	// buffer write operations are potentially slow, a new buffer
	// takes the stage when the current is sent to write
	pool *internal.BufferPool
	// current buffer
	buf *internal.Buffer

	data   chan []byte
	closed bool
	// done blocks all the clients calling to Close() until
	// all writes finish
	done chan struct{}
	// wg is a group of writes in progress
	// done channel won't be closed until wg is done
	wg sync.WaitGroup

	stats Stats
	// client variables
	Writer        io.Writer // writer can be changed at any time
	flushInterval time.Duration
	size          int
}

// NewTickedBuffer wraps a writer with a buffer layer that will write to an underlying writer
// when its buffer is full or a timer ticks
// Call Close to free goroutines, Close blocks until all buffers flush, calling Close and then Write won't panic
func NewTickedBuffer(w io.Writer, bufSize, poolSize, queueSize int, flushInterval time.Duration) *TickedBuffer {
	tb := &TickedBuffer{
		pool:          internal.NewBufferPool(poolSize, bufSize),
		data:          make(chan []byte, queueSize),
		done:          make(chan struct{}),
		flushInterval: flushInterval,
		Writer:        w,
		size:          bufSize,
	}

	// start listening in background
	go tb.listen()

	return tb
}

// ErrBlockingWrite is returned when trying to write on a full channel
var ErrBlockingWrite = errors.New("write was blocking")

// ErrWriteOnClosed is returned when a write is done after closing
var ErrWriteOnClosed = errors.New("write on closed writer")

// Write enqueues the data to be buffered; in order to accomplish a non-blocking write,
// when the buffer channel is full p is discarded and an error is returned
func (tb *TickedBuffer) Write(p []byte) (int, error) {
	if tb.closed {
		return 0, ErrWriteOnClosed
	}
	select {
	case tb.data <- p:
		return len(p), nil
	default:
		atomic.AddInt32(&tb.stats.DroppedWrites, 1)
		return 0, ErrBlockingWrite
	}
}

// Close is concurrent safe and blocks until the remaining data
// in buffer is flushed
func (tb *TickedBuffer) Close() error {
	tb.closed = true
	<-tb.done
	return nil
}

// listen runs a goroutine who acts as a server for all the concurrent writes,
// new goroutines are created to process the writes to the underlying writer
func (tb *TickedBuffer) listen() {
	t := time.NewTicker(tb.flushInterval)
	defer t.Stop()
	// control flag to not flush per tick if a flush is
	// already done by full buffer
	flushedBetweenTicks := false

	tb.buf = tb.getBuffer()
	for {
		// when closed; let the data in the channel be processed
		if tb.closed && len(tb.data) == 0 {
			break
		}
		select {
		case p := <-tb.data:
			// if data is bigger than the buffer write directly to the
			// underlying writer to avoid an unnecessary copy
			if len(p) >= tb.size {
				tb.wg.Add(1)
				go func(p []byte) {
					_, err := tb.Writer.Write(p)
					if err != nil {
						atomic.AddInt32(&tb.stats.FlushErrors, 1)
					}
					tb.wg.Done()
				}(p)
				flushedBetweenTicks = true
			} else {
				// if data is going to outbound the current
				// buffer, flush first
				if len(p) > tb.buf.Available() {
					tb.flush()
					flushedBetweenTicks = true
				}
				tb.buf.Write(p)
			}
		case <-t.C:
			if !flushedBetweenTicks {
				tb.flush()
			} else {
				flushedBetweenTicks = false
			}
		}
	}

	// flush the residual data in buffers
	tb.flush()
	// wait all flushes to finish
	tb.wg.Wait()
	// close done channel to inform clients to stop waiting
	// the remaining buffers to flush
	close(tb.done)
}

func (tb *TickedBuffer) getBuffer() *internal.Buffer {
	b, alloc := tb.pool.Get()
	if alloc {
		tb.stats.BufferAllocs++
	}
	return b
}

// flush writes all the data of the current buffer to the underlying writer
// the buffer used to write is put in background and its sent back to the
// buffer pool when its operation finishes. A new buffer is obtained to continue
// serving incoming writes.
func (tb *TickedBuffer) flush() {
	if tb.buf.Buffered() == 0 {
		return
	}
	buf := tb.buf
	tb.buf = tb.getBuffer()
	tb.wg.Add(1)
	go func() {
		_, err := buf.WriteTo(tb.Writer)
		if err != nil {
			atomic.AddInt32(&tb.stats.FlushErrors, 1)
		}
		tb.pool.Put(buf)
		tb.wg.Done()
	}()
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
	// Writes to our writer that couldn't be attented and were lost in consequence
	DroppedWrites int32
}

// Stats returns a copy of the current writer stats
func (tb *TickedBuffer) Stats() Stats {
	return tb.stats
}
