package ios

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewScannerByDelim(t *testing.T) {
	for _, test := range []struct {
		name           string
		input          io.Reader
		delim          []byte
		flags          ScannerByDelimFlag
		expectedTokens []string
	}{
		{
			name:           "multi-char delim | eof as delim | drop delim",
			input:          strings.NewReader("abc#123##efg####???##xyz##"),
			delim:          []byte("##"),
			flags:          ScannerByDelimFlagEofAsDelim | ScannerByDelimFlagDropDelimInReturn,
			expectedTokens: []string{"abc#123", "efg", "", "???", "xyz"},
		},
		{
			name:           "CR LF delim | eof as delim | include delim",
			input:          strings.NewReader("\r\n\rabc\r"),
			delim:          []byte("\r\n"),
			flags:          ScannerByDelimFlagEofAsDelim | ScannerByDelimFlagIncludeDelimInReturn,
			expectedTokens: []string{"\r\n", "\rabc\r"},
		},
		{
			name:           "empty reader",
			input:          strings.NewReader(""),
			delim:          []byte("*"),
			flags:          ScannerByDelimFlagDefault,
			expectedTokens: []string{},
		},
		{
			name:           "empty token",
			input:          strings.NewReader("*"),
			delim:          []byte("*"),
			flags:          ScannerByDelimFlagEofNotAsDelim | ScannerByDelimFlagDropDelimInReturn,
			expectedTokens: []string{""},
		},
		{
			name:           "trailing newlines",
			input:          strings.NewReader("*\n"),
			delim:          []byte("*"),
			flags:          ScannerByDelimFlagEofAsDelim | ScannerByDelimFlagIncludeDelimInReturn,
			expectedTokens: []string{"*", "\n"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := NewScannerByDelim(test.input, test.delim, test.flags)
			tokens := []string{}
			for s.Scan() {
				tokens = append(tokens, s.Text())
			}
			assert.NoError(t, s.Err())
			assert.Equal(t, test.expectedTokens, tokens)
		})
	}
}

func TestNewScannerByDelim2(t *testing.T) {
	s := NewScannerByDelim2(
		strings.NewReader("abc#123##efg####???##xyz##"),
		[]byte("##"),
		[]byte("?"),
		ScannerByDelimFlagEofAsDelim|ScannerByDelimFlagDropDelimInReturn)
	var tokens []string
	for s.Scan() {
		tokens = append(tokens, s.Text())
	}
	assert.NoError(t, s.Err())
	assert.Equal(t, []string{"abc#123", "efg", "", "???##xyz"}, tokens)
}

func TestNewScannerByDelim3(t *testing.T) {
	buf := make([]byte, 0, 100)
	s := NewScannerByDelim3(
		strings.NewReader("abc#123##efg####???##xyz##"),
		[]byte("##"),
		[]byte("?"),
		ScannerByDelimFlagEofAsDelim|ScannerByDelimFlagDropDelimInReturn,
		buf)
	var tokens []string
	for s.Scan() {
		tokens = append(tokens, s.Text())
	}
	assert.NoError(t, s.Err())
	assert.Equal(t, []string{"abc#123", "efg", "", "???##xyz"}, tokens)
}

var benchmarkInput = strings.Repeat("abc#", 100000)
var benchmarkDelim = []byte("#")
var benchmarkBuf = make([]byte, 1024)

// BenchmarkNewScannerByDelim2-8                            	     500	   2597188 ns/op	    4344 B/op	       5 allocs/op
func BenchmarkNewScannerByDelim2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewScannerByDelim2(
			strings.NewReader(benchmarkInput),
			benchmarkDelim,
			nil,
			ScannerByDelimFlagEofAsDelim|ScannerByDelimFlagDropDelimInReturn)
		for s.Scan() {
			token := s.Bytes()
			if len(token) != 3 || token[0] != 'a' || token[1] != 'b' || token[2] != 'c' {
				b.FailNow()
			}
		}
		if s.Err() != nil {
			b.FailNow()
		}
	}
}

// BenchmarkNewScannerByDelim3-8                            	     500	   2585825 ns/op	     242 B/op	       3 allocs/op
func BenchmarkNewScannerByDelim3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewScannerByDelim3(
			strings.NewReader(benchmarkInput),
			benchmarkDelim,
			nil,
			ScannerByDelimFlagEofAsDelim|ScannerByDelimFlagDropDelimInReturn,
			benchmarkBuf)
		for s.Scan() {
			token := s.Bytes()
			if len(token) != 3 || token[0] != 'a' || token[1] != 'b' || token[2] != 'c' {
				b.FailNow()
			}
		}
		if s.Err() != nil {
			b.FailNow()
		}
	}
}
