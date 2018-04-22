package asyncio

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testWriter struct {
	writes int64
	bytes  int64
}

func (w *testWriter) Write(p []byte) (int, error) {
	atomic.AddInt64(&w.writes, 1)
	atomic.AddInt64(&w.bytes, int64(len(p))
	return len(p), nil
}

func TestConcurrentWrites(t *testing.T) {
	concurrency := 100
	size := 1024
	tw := &testWriter{}
	tb := NewTickedBuffer(tw, size, 2, concurrency, time.Second)

	p := make([]byte, size)
	wg := sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			n, err := tb.Write(p)
			if err != nil {
				t.Fatalf(err.Error())
			}
			if n != size {
				t.Fatalf("writed %v bytes, exepcted: %v", n, size)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	tb.Close()

	t.Logf("testWriter: %v", *tw)
	t.Logf("TickedBuffer stats: %v", tb.Stats())

	if tw.writes != int64(concurrency) {
		t.Errorf("test writer writes: %v, actual writes: %v", tw.writes, concurrency)
	}
	if tw.bytes != int64(size*concurrency) {
		t.Errorf("test writer bytes: %v, actual bytes: %v", tw.bytes, size*concurrency)
	}
}

// TestTicks expects the TickedBuffer to flush exclusively by ticks
func TestTicks(t *testing.T) {
	tickInterval := time.Millisecond * 200
	iterations := 10
	size := 1020
	block := size / iterations

	tw := &testWriter{}
	tb := NewTickedBuffer(tw, size+1, 2, 1, tickInterval)

	p := make([]byte, block)
	for i := 0; i < iterations; i++ {
		n, err := tb.Write(p)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if n != block {
			t.Fatalf("writed %v bytes, exepcted: %v", n, block)
		}
		time.Sleep(tickInterval)
	}
	tb.Close()

	t.Logf("testWriter: %v", *tw)
	t.Logf("TickedBuffer stats: %v", tb.Stats())

	if tw.writes != int64(iterations) {
		t.Errorf("test writer writes: %v, actual writes: %v", tw.writes, iterations)
	}
	if tw.bytes != int64(size) {
		t.Errorf("test writer bytes: %v, actual bytes: %v", tw.bytes, size)
	}
}

func BenchmarkConcurrentWrites(b *testing.B) {
	size := 1024
	tw := &testWriter{}
	tb := NewTickedBuffer(tw, size*100, b.N/100, b.N, time.Second*2)

	p := make([]byte, size)
	for n := 0; n < b.N; n++ {
		tb.Write(p)
	}

	tb.Close()
	b.Logf("Writer stats: %v Dropped writes: %v, buffers: %v", tw, tb.stats.DroppedWrites, tb.stats.BufferAllocs)
}
