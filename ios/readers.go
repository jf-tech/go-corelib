package iohelper

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
)

// LineNumReportingCsvReader wraps std lib `*csv.Reader` and exposes the current line number.
type LineNumReportingCsvReader struct {
	*csv.Reader
	numLineFieldName string
}

// LineNum returns the current line number
func (r *LineNumReportingCsvReader) LineNum() int {
	// Given the request to add line num report api on csv.Reader is denied
	// per https://github.com/golang/go/issues/26679, let's use this hack to get
	// the line number for our internal use.
	value := reflect.ValueOf(*r).FieldByName(r.numLineFieldName)
	if !value.IsValid() {
		panic(fmt.Sprintf(
			"unable to get '%s' from csv.Reader, has csv.Reader been changed/upgraded?", r.numLineFieldName))
	}
	return int(value.Int())
}

// NewLineNumReportingCsvReader creates a new `*LineNumReportingCsvReader`.
func NewLineNumReportingCsvReader(r io.Reader) *LineNumReportingCsvReader {
	return &LineNumReportingCsvReader{csv.NewReader(r), "numLine"}
}

// BytesReplacingReader allows transparent replacement of a given token during read operation.
type BytesReplacingReader struct {
	r          io.Reader
	search     []byte
	searchLen  int
	replace    []byte
	replaceLen int
	lenDelta   int // = replaceLen - searchLen. can be negative
	err        error
	buf        []byte
	buf0, buf1 int // buf[0:buf0]: bytes already processed; buf[buf0:buf1] bytes read in but not yet processed.
	max        int // because we need to replace 'search' with 'replace', this marks the max bytes we can read into buf
}

const defaultBufSize = int(4096)

// NewBytesReplacingReader creates a new `*BytesReplacingReader`.
// `search` cannot be nil/empty. `replace` can.
func NewBytesReplacingReader(r io.Reader, search, replace []byte) *BytesReplacingReader {
	return (&BytesReplacingReader{}).Reset(r, search, replace)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Reset allows reuse of a previous allocated `*BytesReplacingReader` for buf allocation optimization.
// `search` cannot be nil/empty. `replace` can.
func (r *BytesReplacingReader) Reset(r1 io.Reader, search1, replace1 []byte) *BytesReplacingReader {
	if r1 == nil {
		panic("io.Reader cannot be nil")
	}
	if len(search1) == 0 {
		panic("search token cannot be nil/empty")
	}
	r.r = r1
	r.search = search1
	r.searchLen = len(search1)
	r.replace = replace1
	r.replaceLen = len(replace1)
	r.lenDelta = r.replaceLen - r.searchLen // could be negative
	r.err = nil
	bufSize := max(defaultBufSize, max(r.searchLen, r.replaceLen))
	if r.buf == nil || len(r.buf) < bufSize {
		r.buf = make([]byte, bufSize)
	}
	r.buf0 = 0
	r.buf1 = 0
	r.max = len(r.buf)
	if r.searchLen < r.replaceLen {
		// If len(search) < len(replace), then we have to assume the worst case:
		// what's the max bound value such that if we have consecutive 'search' filling up
		// the buf up to buf[:max], and all of them are placed with 'replace', and the final
		// result won't end up exceed the len(buf)?
		r.max = (len(r.buf) / r.replaceLen) * r.searchLen
	}
	return r
}

// Read implements the `io.Reader` interface.
func (r *BytesReplacingReader) Read(p []byte) (int, error) {
	n := 0
	for {
		if r.buf0 > 0 {
			n = copy(p, r.buf[0:r.buf0])
			r.buf0 -= n
			r.buf1 -= n
			if r.buf1 == 0 && r.err != nil {
				return n, r.err
			}
			copy(r.buf, r.buf[n:r.buf1+n])
			return n, nil
		} else if r.err != nil {
			return 0, r.err
		}

		n, r.err = r.r.Read(r.buf[r.buf1:r.max])
		if n > 0 {
			r.buf1 += n
			for {
				index := bytes.Index(r.buf[r.buf0:r.buf1], r.search)
				if index < 0 {
					r.buf0 = max(r.buf0, r.buf1-r.searchLen+1)
					break
				}
				index += r.buf0
				copy(r.buf[index+r.replaceLen:r.buf1+r.lenDelta], r.buf[index+r.searchLen:r.buf1])
				copy(r.buf[index:index+r.replaceLen], r.replace)
				r.buf0 = index + r.replaceLen
				r.buf1 += r.lenDelta
			}
		}
		if r.err != nil {
			r.buf0 = r.buf1
		}
	}
}

// ReadLine reads in a single line from a bufio.Reader.
func ReadLine(r *bufio.Reader) (string, error) {
	// Turns out even with various bufio.Reader.Read???() and bufio.Scanner, there is not simple clean
	// way of reading a single text line in:
	// - bufio.ReadSlice('\n') doesn't have '\r' dropping. We want a line returned without neither '\n' nor '\r'.
	// - bufio.ReadLine() drops '\r' and '\n', but has a fixed buf so may be unable to read a whole line in one call.
	// - bufio.ReadBytes no buf size issue, but doesn't offer '\r' and '\n' cleanup.
	// - bufio.ReadString essentially the same as bufio.ReadBytes.
	// - bufio.Scanner deals with '\r' and '\n' but has fixed buf issue.
	// Oh, come on!!
	//
	// Also found net/textproto's Reader.ReadLine() which meets all the requirements. But to use it
	// we need to create yet another type of Reader (net.textproto.Reader), as if the
	// io.Reader -> bufio.Reader isn't enough for us. So decided instead, just shamelessly copy
	// net.textproto.Reader.ReadLine() here, credit goes to
	// https://github.com/golang/go/blob/master/src/net/textproto/reader.go. However its test code
	// coverage is lacking, so create all the new test cases for this ReadLine implementation copy.
	var line []byte
	for {
		l, more, err := r.ReadLine()
		if err != nil {
			return "", err
		}
		// Avoid the copy if the first call produced a full line.
		if line == nil && !more {
			return string(l), nil
		}
		line = append(line, l...)
		if !more {
			break
		}
	}
	return string(line), nil
}

const bom = '\uFEFF'

// StripBOM returns a new io.Reader that, if needed, strips away the BOM (byte order marker) of
// the input io.Reader.
func StripBOM(reader io.Reader) (io.Reader, error) {
	br := bufio.NewReader(reader)
	r, _, err := br.ReadRune()
	switch {
	case err == io.EOF:
		// This is to handle empty file, can't call UnreadRune(), will meet ErrInvalidUnreadRune as
		// b.lastRuneSize is -1. So simply reset buffer io.
		br.Reset(reader)
		return br, nil
	case err != nil:
		return nil, err
	case r == bom:
		return br, nil
	default:
		// Here we shouldn't meet any error during unread rune.
		_ = br.UnreadRune()
		return br, nil
	}
}

// LineCountingReader wraps an io.Reader and reports currently which line the reader is at. Note
// the LineAt starts a 1.
type LineCountingReader struct {
	r    io.Reader
	line int
}

// Read implements the `io.Reader` interface.
func (r *LineCountingReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	// We can do this byte search and count because '\n' < utf8.RuneSelf.
	for i := 0; i < n; i++ {
		// On Windows a line ends with "\r\n" on Linux/Unix/MacOS it ends
		// with "\n". Either way, searching for "\n" for is good enough for
		// line counting purpose.
		if p[i] == '\n' {
			r.line++
		}
	}
	return n, err
}

// AtLine returns the current line number. Note it starts with 1.
func (r *LineCountingReader) AtLine() int {
	return r.line
}

// NewLineCountingReader creates new LineCountingReader wrapping around an input io.Reader.
func NewLineCountingReader(r io.Reader) *LineCountingReader {
	return &LineCountingReader{r: r, line: 1}
}
