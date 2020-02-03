/* Package easywriter mirrors and extends the bufio.Writer interface, but delays error handling.

Instead of having to check the error on every call, you can write a few parts and
then check for errors once you completed a part.
*/
package easywriter

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

const (
	defaultBufSize = 4096
)

// Writer is an io.Writer with many convenience methods that allow delayed error
// checking.
type Writer struct {
	bw      *bufio.Writer
	err     error
	tmp     []byte
	tmpdata [64]byte // prevents heap alloc, fits 64 bit number formatted as binary
}

// New constructs a Writer from an io.Writer. It wraps it in a bufio.Writer,
// unless the passed in value already is a *bufio.Writer.
func New(w io.Writer) *Writer {
	if bw, ok := w.(*bufio.Writer); ok {
		// Optimization if w already is a bufio.Writer
		return FromBufIOWriter(bw)
	}
	return NewSize(w, defaultBufSize)
}

// New constructs a Writer from an io.Writer. It wraps it in a bufio.Writer with
// given buffer size.
func NewSize(w io.Writer, size int) *Writer {
	bw := bufio.NewWriterSize(w, size)
	return FromBufIOWriter(bw)
}

// New constructs a Writer from a bufio.Writer.
func FromBufIOWriter(bw *bufio.Writer) *Writer {
	w := Writer{
		bw: bw,
	}
	w.tmp = w.tmpdata[:]
	return &w
}

// Err returns the current error, if any. Reading the error does not reset it.
func (b *Writer) Err() error {
	return b.err
}

// ResetErr reset the error to nil. You should never need this.
func (b *Writer) ResetErr() {
	b.err = nil
}

// bufio.Writer interface, but with error stripped

// Available returns the number of bytes still available in the underlying buffer.
func (b *Writer) Available() int {
	return b.bw.Available()
}

// Buffered returns the number of bytes buffered in the underlying buffer.
func (b *Writer) Buffered() int {
	return b.bw.Buffered()

}

// Size returns the size of the underlying buffer.
func (b *Writer) Size() int {
	return b.bw.Size()
}

// Flush still returns an error, to keep satisfying the io.Flusher interface,
// and as a natural point to check errors.
// FlushInterim does the same without returning an error.
func (b *Writer) Flush() error {
	if b.err != nil {
		return b.err
	}
	b.err = b.bw.Flush()
	return b.err
}

// FlushInterim performs a Flush without returning an error.
func (b *Writer) FlushInterim() {
	if b.err != nil {
		return
	}
	b.err = b.bw.Flush()
}

// ReadFrom still returns an error, to keep satisfying the io.ReaderFrom interface.
// ReadBytesFrom does the same without returning an error.
func (b *Writer) ReadFrom(r io.Reader) (n int64, err error) {
	if b.err != nil {
		return 0, b.err
	}
	n, b.err = b.bw.ReadFrom(r)
	return n, b.err
}

// ReadBytesFrom is a version of ReadFrom that does not return an error.
func (b *Writer) ReadBytesFrom(r io.Reader) (n int64) {
	if b.err != nil {
		return 0
	}
	n, b.err = b.bw.ReadFrom(r)
	return n
}

// Write still returns an error, to keep satisfying the io.Writer interface.
// WriteBytes does the same without returning an error.
func (b *Writer) Write(p []byte) (nn int, err error) {
	if b.err != nil {
		return 0, b.err
	}
	nn, b.err = b.bw.Write(p)
	return nn, b.err
}

// WriteBytes writes bytes without returning an error.
func (b *Writer) WriteBytes(p []byte) (nn int) {
	if b.err != nil {
		return 0
	}
	nn, b.err = b.bw.Write(p)
	return nn
}

// WriteByte writes a single byte without returning an error.
func (b *Writer) WriteByte(c byte) {
	if b.err != nil {
		return
	}
	b.err = b.bw.WriteByte(c)
}

// WriteByte writes a single rune without returning an error.
func (b *Writer) WriteRune(r rune) (size int) {
	if b.err != nil {
		return
	}
	size, b.err = b.bw.WriteRune(r)
	return size
}

// WriteString writes a string without returning an error.
func (b *Writer) WriteString(s string) (n int) {
	if b.err != nil {
		return 0
	}
	n, b.err = b.bw.WriteString(s)
	return n
}

// Additional useful methods

// WriteString writes an int in decimal text representation and returns the
// number of bytes written.
func (b *Writer) WriteDecimal(num int) (n int) {
	return b.WriteNumber64(int64(num), 10)
}

// WriteString writes an int in text representation with given base and returns the
// number of bytes written.
func (b *Writer) WriteNumber(num, base int) (n int) {
	return b.WriteNumber64(int64(num), base)
}

// WriteString writes a uint in text representation with given base and returns the
// number of bytes written.
func (b *Writer) WriteUnsignedNumber(num uint, base int) (n int) {
	return b.WriteUnsignedNumber64(uint64(num), base)
}

// WriteString writes an int64 in text representation with given base and returns the
// number of bytes written.
func (b *Writer) WriteNumber64(num int64, base int) (n int) {
	if b.err != nil {
		return 0
	}
	t := b.tmp[:0]
	t = strconv.AppendInt(t, num, base)
	n, b.err = b.bw.Write(t)
	return n
}

// WriteString writes a uint64 in text representation with given base and returns the
// number of bytes written.
func (b *Writer) WriteUnsignedNumber64(num uint64, base int) (n int) {
	if b.err != nil {
		return 0
	}
	t := b.tmp[:0]
	t = strconv.AppendUint(t, num, base)
	n, b.err = b.bw.Write(t)
	return n
}

// Printf writes a string with given format and returns the number of bytes written.
func (b *Writer) Printf(format string, a ...interface{}) (n int) {
	if b.err != nil {
		return 0
	}
	n, b.err = fmt.Fprintf(b.bw, format, a...)
	return n
}

// Println writes space separated values plus a newline and returns the number of bytes written.
func (b *Writer) Println(a ...interface{}) (n int) {
	if b.err != nil {
		return 0
	}
	n, b.err = fmt.Fprintln(b.bw, a...)
	return n
}

// Println writes space separated values without a newline and returns the number of bytes written.
func (b *Writer) Print(a ...interface{}) (n int) {
	if b.err != nil {
		return 0
	}
	n, b.err = fmt.Fprint(b.bw, a...)
	return n
}

// WriteUint16LE writes given value in binary with Little Endian order.
func (b *Writer) WriteUint16LE(v uint16) {
	if b.err != nil {
		return
	}
	t := b.tmp[:2]
	binary.LittleEndian.PutUint16(t, v)
	_, b.err = b.bw.Write(t)
}

// WriteUint32LE writes given value in binary with Little Endian order.
func (b *Writer) WriteUint32LE(v uint32) {
	if b.err != nil {
		return
	}
	t := b.tmp[:4]
	binary.LittleEndian.PutUint32(t, v)
	_, b.err = b.bw.Write(t)
}

// WriteUint64LE writes given value in binary with Little Endian order.
func (b *Writer) WriteUint64LE(v uint64) {
	if b.err != nil {
		return
	}
	t := b.tmp[:8]
	binary.LittleEndian.PutUint64(t, v)
	_, b.err = b.bw.Write(t)
}

// WriteUint16BE writes given value in binary with Big Endian order.
func (b *Writer) WriteUint16BE(v uint16) {
	if b.err != nil {
		return
	}
	t := b.tmp[:2]
	binary.BigEndian.PutUint16(t, v)
	_, b.err = b.bw.Write(t)
}

// WriteUint32BE writes given value in binary with Big Endian order.
func (b *Writer) WriteUint32BE(v uint32) {
	if b.err != nil {
		return
	}
	t := b.tmp[:4]
	binary.BigEndian.PutUint32(t, v)
	_, b.err = b.bw.Write(t)
}

// WriteUint64BE writes given value in binary with Big Endian order.
func (b *Writer) WriteUint64BE(v uint64) {
	if b.err != nil {
		return
	}
	t := b.tmp[:8]
	binary.BigEndian.PutUint64(t, v)
	_, b.err = b.bw.Write(t)
}
