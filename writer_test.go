package easywriter

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	w := New(ioutil.Discard)
	w.WriteString("foo")
	if err := w.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestNewWriterSize(t *testing.T) {
	w := NewSize(ioutil.Discard, 123)
	if w.Size() != 123 {
		t.Fatal("size not used", w.Size())
	}
	w.WriteString("foo")
	if err := w.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestNewWriterBufIO(t *testing.T) {
	bi := bufio.NewWriterSize(ioutil.Discard, 42)
	w := FromBufIOWriter(bi)
	if w.Size() != 42 {
		t.Fatal("size not used", w.Size())
	}
	w.WriteString("foo")
	if err := w.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestWriter_Size_Buffered_Available_Flush(t *testing.T) {
	w := NewSize(ioutil.Discard, 100)
	if w.Size() != 100 {
		t.Fatal("size not used", w.Size())
	}
	if w.Buffered() != 0 {
		t.Fatal("buffered not zero", w.Buffered())
	}
	if w.Available() != 100 {
		t.Fatal("available not full size", w.Available())
	}
	w.WriteString("foo")
	if w.Size() != 100 {
		t.Fatal("size changed", w.Size())
	}
	if w.Buffered() != 3 {
		t.Fatal("buffered wrong", w.Buffered())
	}
	if w.Available() != 97 {
		t.Fatal("available wrong", w.Available())
	}
	w.FlushInterim()
	if w.Buffered() != 0 {
		t.Fatal("buffered not zero after flush", w.Buffered())
	}
	if w.Available() != 100 {
		t.Fatal("available not full size after flush", w.Available())
	}
}

var testErr = errors.New("test error")

type WriterWithError struct{}

func (w WriterWithError) Write([]byte) (int, error) {
	return 0, testErr
}

func TestWriter_Error(t *testing.T) {
	w := New(WriterWithError{})
	w.WriteString("foo")
	w.FlushInterim()
	if err := w.Err(); err != testErr {
		t.Fatal("Did not return expected error, but", err)
	}
	if err := w.Err(); err != testErr {
		t.Fatal("Did not return expected error on second call, but", err)
	}
	w.ResetErr()
	if err := w.Err(); err != nil {
		t.Fatal("Err not reset, got:", err)
	}
}

var expectedText = `
Hello, world
Write
😀

123
123
1111011
7b
-123
123
-100000000000000000000000000000000000000000000000000000000000000
1000000000000000000000000000000000000000000000000000000000000000

The number is 042
1 2 3 foo
1 2 3no space nor newline
injected from reader
`

// TODO: convert into an example
func TestWriter_Write_Text(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	w := New(buf)

	w.WriteByte('\n')
	w.WriteString("Hello, world\n")
	w.WriteBytes([]byte("Write\n"))
	w.WriteRune(0x1F600)
	w.WriteByte('\n')
	w.WriteByte('\n')

	w.WriteDecimal(123)
	w.WriteByte('\n')
	w.WriteNumber(123, 10)
	w.WriteByte('\n')
	w.WriteNumber(123, 2)
	w.WriteByte('\n')
	w.WriteNumber(123, 16)
	w.WriteByte('\n')
	w.WriteNumber(-123, 10)
	w.WriteByte('\n')
	w.WriteUnsignedNumber(123, 10)
	w.WriteByte('\n')
	w.WriteNumber64(-1<<62, 2)
	w.WriteByte('\n')
	w.WriteUnsignedNumber64(1<<63, 2)
	w.WriteByte('\n')
	w.WriteByte('\n')

	w.Printf("The number is %03d\n", 42)
	w.Println(1, 2, 3, "foo")
	w.Print(1, 2, 3, "no space nor newline")
	w.WriteByte('\n')

	part := bytes.NewBuffer([]byte("injected from reader\n"))
	w.ReadBytesFrom(part)

	w.FlushInterim()
	if err := w.Err(); err != nil {
		t.Fatal("Unexpected error:", err)
	}

	got := string(buf.Bytes())
	if got != expectedText {
		t.Fatalf("Expected:\n---\n%s---\nGot:\n---\n%s---\n", expectedText, got)
	}
}

var expectedBinary = `
A.
B...
C.......
.D
...E
.......F
`

func TestWriter_Write_Binary(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	w := New(buf)

	w.WriteByte('\n')
	w.WriteUint16LE(65)
	w.WriteByte('\n')
	w.WriteUint32LE(66)
	w.WriteByte('\n')
	w.WriteUint64LE(67)
	w.WriteByte('\n')
	w.WriteUint16BE(68)
	w.WriteByte('\n')
	w.WriteUint32BE(69)
	w.WriteByte('\n')
	w.WriteUint64BE(70)
	w.WriteByte('\n')

	w.FlushInterim()
	if err := w.Err(); err != nil {
		t.Fatal("Unexpected error:", err)
	}

	got := strings.ReplaceAll(string(buf.Bytes()), "\000", ".")
	if got != expectedBinary {
		t.Fatalf("Expected:\n---\n%s---\nGot:\n---\n%s---\n", expectedText, got)
	}
}

func TestWriter_Write_Pending_Error(t *testing.T) {
	// Test that we do not write anything after an error

	buf := bytes.NewBuffer(nil)
	w := New(buf)
	w.err = testErr
	if err := w.Err(); err != testErr {
		t.Fatal("Not the error we set:", err)
	}

	w.WriteByte('\n')
	w.WriteString("Hello, world\n")
	w.WriteBytes([]byte("Write\n"))
	w.WriteRune(0x1F600)
	w.WriteDecimal(123)
	w.WriteNumber(234, 10)
	w.WriteUnsignedNumber(345, 10)
	w.Printf("The number is %03d\n", 42)
	w.Println(1, 2, 3)
	w.Print(1, 2, 3, "no space nor newline")

	part := bytes.NewBuffer([]byte("injected from reader\n"))
	w.ReadBytesFrom(part)

	w.WriteUint16LE(65)
	w.WriteUint32LE(66)
	w.WriteUint64LE(67)
	w.WriteUint16BE(68)
	w.WriteUint32BE(69)
	w.WriteUint64BE(70)

	w.FlushInterim()
	if err := w.Err(); err != testErr {
		t.Fatal("Not the error we set:", err)
	}

	got := string(buf.Bytes())
	if got != "" {
		t.Fatalf("Expected empty string, but got: %q", got)
	}
}

func BenchmarkWriter_WriteByte(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteByte('\n')
	}
}

func BenchmarkWriter_WriteByte_underlying(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(1)
	bw := w.bw
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bw.WriteByte('\n')
	}
}

func BenchmarkWriter_WriteString_13b(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(13)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteString("Hello, world\n")
	}
}

func BenchmarkWriter_WriteBytes_13b(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	data := []byte("Hello, world\n")
	b.ReportAllocs()
	b.SetBytes(13)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteBytes(data)
	}
}

func BenchmarkWriter_WriteBytes_13b_underlying(b *testing.B) {
	// For comparison, skip our wrapper and write directly to the bufio.Buffer
	w := New(ioutil.Discard)
	data := []byte("Hello, world\n")
	b.ReportAllocs()
	b.SetBytes(13)
	bw := w.bw
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bw.Write(data)
	}
}

func BenchmarkWriter_WriteRune(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(int64(w.WriteRune(0x1F600)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteRune(0x1F600)
	}
}

func BenchmarkWriter_WriteDecimal(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteDecimal(123)
	}
}

func BenchmarkWriter_WriteNumber(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteNumber(123, 10)
	}
}

func BenchmarkWriter_WriteUnsignedNumber(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteUnsignedNumber(123, 10)
	}
}

func BenchmarkWriter_WriteNumber64(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteNumber64(123, 10)
	}
}

func BenchmarkWriter_WriteUnsignedNumber64(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteUnsignedNumber64(123, 10)
	}
}

func BenchmarkWriter_WriteUnsignedNumber64_binary_allbits(b *testing.B) {
	// This requires tmp to be 64 bytes to not allocate memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteUnsignedNumber64(1<<63, 2)
	}
}

func BenchmarkWriter_Printf_number(b *testing.B) {
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Printf("%d", 123)
	}
}

func BenchmarkWriter_Print_number(b *testing.B) {
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Print(123)
	}
}

func BenchmarkWriter_Println_number(b *testing.B) {
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Println(123)
	}
}

func BenchmarkWriter_WriteUint16LE(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteUint16LE(42)
	}
}

func BenchmarkWriter_WriteUint32LE(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteUint32LE(42)
	}
}

func BenchmarkWriter_WriteUint64LE(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteUint64LE(42)
	}
}

func BenchmarkWriter_WriteUint16BE(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteUint16BE(42)
	}
}

func BenchmarkWriter_WriteUint32BE(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteUint32BE(42)
	}
}

func BenchmarkWriter_WriteUint64BE(b *testing.B) {
	// This subset is expected not to alloc any memory
	w := New(ioutil.Discard)
	b.ReportAllocs()
	b.SetBytes(8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteUint64BE(42)
	}
}
