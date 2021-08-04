package ios

import (
	"bufio"
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

// ByteReadLine reads in a single line from a bufio.Reader and returns it in []byte.
// Note the returned []byte may be pointing directly into the bufio.Reader, so assume
// the returned []byte will be invalidated and shouldn't be used upon next ByteReadLine
// call.
func ByteReadLine(r *bufio.Reader) ([]byte, error) {
	// We want to read a line delimited by '\n' in cleanly with trailing '\r' '\n' dropped. Turns out
	// none of the bufio.Reader.Read???() and bufio.Scanner can do that:
	// - bufio.ReadSlice('\n') doesn't have '\r' dropping. Also it may need multiple calls to get a single (long) line.
	// - bufio.ReadLine() drops '\r' and '\n', but may need multiple calls to get a single (long) line.
	// - bufio.ReadBytes doesn't need multiple calls, but doesn't offer '\r' and '\n' cleanup. Also tons of allocations.
	// - bufio.ReadString essentially the same as bufio.ReadBytes.
	// - bufio.Scanner deals with '\r' and '\n' but may need multiple calls to get a single (long) line.
	//
	// Found net/textproto's Reader.ReadLine() which meets all the requirements. But it would become another
	// dependency. So decided just shamelessly copy net.textproto.Reader.ReadLine() here, credit goes to
	// https://github.com/golang/go/blob/master/src/net/textproto/reader.go. Its test code coverage is lacking,
	// so create all the new test cases for this ReadLine implementation copy.
	var line []byte
	for {
		l, more, err := r.ReadLine()
		if err != nil {
			return nil, err
		}
		// Avoid the copy if the first call produced a full line.
		if line == nil && !more {
			line = l
			goto returnLine
		}
		line = append(line, l...)
		if !more {
			break
		}
	}
returnLine:
	if len(line) == 0 {
		return nil, nil
	}
	return line, nil
}

// ReadLine reads in a single line from a bufio.Reader and returns it in string.
func ReadLine(r *bufio.Reader) (string, error) {
	b, err := ByteReadLine(r)
	return string(b), err
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
