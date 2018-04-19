package internal

import (
	"fmt"
	"io"
)

// Buffer is a sized Buffer which implement write methods
// differs from bytes.Buffer in that it can't grow or shrink,
// this permit us to have control over available bytes.
// bufio.Writer could suit our use but this implementation is stricter
// on when to write; also permits us to change the underlying writer making
// a pool of Buffer easier to manage.
type Buffer struct {
	buf []byte
	n   int
}

func newBuffer(size int) *Buffer {
	return &Buffer{buf: make([]byte, size)}
}

// Buffered returns the size of the data writen in the buffer
func (b *Buffer) Buffered() int {
	return b.n
}

// Available returns the writable space in the buffer
func (b *Buffer) Available() int {
	return cap(b.buf) - b.n
}

// Write copies p into the internal buffer,
// an error is returned if p is bigger than the available memory
func (b *Buffer) Write(p []byte) (int, error) {
	// make sure that this will never happen
	if len(p) > b.Available() {
		return 0, fmt.Errorf("buffer write error: buffer has %v mem available and p has length %v", b.Available(), len(p))
	}
	n := copy(b.buf[b.n:], p)
	if n != len(p) {
		return 0, fmt.Errorf("buffer write error: p has length %v and %v were writen", len(p), n)
	}
	b.n += n
	return n, nil
}

// WriteTo flushes all the buffer data into a writer
func (b *Buffer) WriteTo(w io.Writer) (n int64, err error) {
	m, err := w.Write(b.buf[:b.n])
	if m > b.n {
		return n, fmt.Errorf("writed more bytes than the buffered")
	}
	n = int64(m)
	if err != nil {
		return n, err
	}
	// all bytes should have been written, by definition of
	// Write method in io.Writer
	if m != b.n {
		return n, io.ErrShortWrite
	}

	// buffer is now empty; reset.
	b.Reset()
	return
}

// Reset sets the buffer as empty
func (b *Buffer) Reset() {
	b.n = 0
}
