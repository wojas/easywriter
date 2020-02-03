# Go file writing with delayed error handling

[![Documentation](https://godoc.org/github.com/wojas/easywriter?status.svg)](http://godoc.org/github.com/wojas/easywriter)

Package easywriter mirrors the bufio.Reader interface (it currently supports the Go 1.13 methods), but with delayed error handling. Instead of having to check the error on every call, you can write a few parts and then check for errors once you completed a part.

Example:

```go
w := easywriter.New(f)
w.WriteString("Hello, world\n")
w.WriteDecimal(123)
w.WriteByte('\n')
w.WriteRune(0x1F600)
w.WriteUint32BE(42)
// etc
if err := w.Flush(); err != nil {}
    return err
}
```

The overhead added by this wrapper is negligible (~1ns) and none of the methods allocate heap memory.


## Advantages over the standard library

Consider this example with error checking on every call:

```go
w := bufio.NewWriter(f)
if _, err := w.WriteString("Hello, world\n"); err != nil {
    return err
}
if _, err := w.Write([]byte{'f', 'o', 'o'}); err != nil {
    return err
}
if err := w.WriteByte('\n'); err != nil {
    return err
}
if err := w.Flush(); err != nil {}
    return err
}
```

This is just painful to read and to write. The standard package does allow you to skip
the first few checks:

> If an error occurs writing to a Writer, no more data will be accepted and all subsequent writes, and Flush, will return the error.

Let's try this:

```go
w := bufio.NewWriter(f)
w.WriteString("Hello, world\n")
w.Write([]byte{'f', 'o', 'o'})
w.WriteByte('\n')
if err := w.Flush(); err != nil {}
    return err
}
```

Thats look a lot better, just like this easywriter package! Unfortunately, now all modern IDEs and lint tools will be shouting at you for implicitly ignoring errors in the first few calls...

Let's try to silence them:

```go
w := bufio.NewWriter(f)
// If an error occurs writing to a Writer, no more data will be accepted and 
// all subsequent writes, and Flush, will return the error.
_, _ = w.WriteString("Hello, world\n")
_, _ = w.Write([]byte{'f', 'o', 'o'})
_ = w.WriteByte('\n')
if err := w.Flush(); err != nil {}
    return err
}
```

This gets messy again and all the explicit error silencing will raise questions during a code review, basically requiring a comment to explain why you are ignoring the errors.

The same code with easywriter looks like this:

```go
w := easywriter.New(f)
w.WriteString("Hello, world\n")
w.WriteBytes([]byte{'f', 'o', 'o'})
w.WriteByte('\n')
if err := w.Flush(); err != nil {}
    return err
}
```

A lot cleaner and no warnings from your IDE and linter. Note that we use `WriteBytes` here: easywriter still keeps a `Write` method that returns an error to satisfy the `io.Writer` interface.


## Benchmarks

These benchmarks write to `io.Discard` through `bufio.Writer` and easywriter:

```
goos: darwin
goarch: amd64
pkg: github.com/wojas/easywriter
BenchmarkWriter_WriteByte-20                               	324578672	         3.70 ns/op	 270.53 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteByte_underlying-20                    	385809908	         3.10 ns/op	 323.05 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteString_13b-20                         	184706305	         6.68 ns/op	1946.43 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteBytes_13b-20                          	169732112	         7.06 ns/op	1841.33 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteBytes_13b_underlying-20               	183907147	         6.49 ns/op	2003.32 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteRune-20                               	166844498	         7.13 ns/op	 560.76 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteDecimal-20                            	58019652	        22.0 ns/op	 136.41 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteNumber-20                             	55473448	        21.5 ns/op	 139.73 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteUnsignedNumber-20                     	55126424	        22.1 ns/op	 136.04 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteNumber64-20                           	56523331	        21.7 ns/op	 137.99 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteUnsignedNumber64-20                   	58061304	        22.4 ns/op	 133.97 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteUnsignedNumber64_binary_allbits-20    	14155111	        85.6 ns/op	 747.80 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_Printf_number-20                           	21304226	        55.8 ns/op	  53.74 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_Print_number-20                            	22473684	        57.1 ns/op	  52.52 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_Println_number-20                          	21757617	        54.8 ns/op	  54.71 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteUint16LE-20                           	147842317	         8.17 ns/op	 244.90 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteUint32LE-20                           	154393246	         7.79 ns/op	 513.66 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteUint64LE-20                           	149820724	         7.97 ns/op	1003.62 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteUint16BE-20                           	149114833	         8.03 ns/op	 249.16 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteUint32BE-20                           	151594344	         7.93 ns/op	 504.10 MB/s	       0 B/op	       0 allocs/op
BenchmarkWriter_WriteUint64BE-20                           	134773581	         8.90 ns/op	 898.52 MB/s	       0 B/op	       0 allocs/op
PASS
```
