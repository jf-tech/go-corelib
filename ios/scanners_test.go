package ios

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/go-corelib/strs"
)

func TestNewScannerByDelim(t *testing.T) {
	for _, test := range []struct {
		name           string
		input          io.Reader
		delim          string
		flags          ScannerByDelimFlag
		expectedTokens []string
	}{
		{
			name:           "multi-char delim | eof as delim | drop delim",
			input:          strings.NewReader("abc#123##efg####???##xyz##"),
			delim:          "##",
			flags:          ScannerByDelimFlagEofAsDelim | ScannerByDelimFlagDropDelimInReturn,
			expectedTokens: []string{"abc#123", "efg", "", "???", "xyz"},
		},
		{
			name:           "CR LF delim | eof as delim | include delim",
			input:          strings.NewReader("\r\n\rabc\r"),
			delim:          "\r\n",
			flags:          ScannerByDelimFlagEofAsDelim | ScannerByDelimFlagIncludeDelimInReturn,
			expectedTokens: []string{"\r\n", "\rabc\r"},
		},
		{
			name:           "empty reader",
			input:          strings.NewReader(""),
			delim:          "*",
			flags:          ScannerByDelimFlagDefault,
			expectedTokens: []string{},
		},
		{
			name:           "empty token",
			input:          strings.NewReader("*"),
			delim:          "*",
			flags:          ScannerByDelimFlagEofNotAsDelim | ScannerByDelimFlagDropDelimInReturn,
			expectedTokens: []string{""},
		},
		{
			name:           "trailing newlines",
			input:          strings.NewReader("*\n"),
			delim:          "*",
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
		"##",
		strs.RunePtr('?'),
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
		"##",
		strs.RunePtr('?'),
		ScannerByDelimFlagEofAsDelim|ScannerByDelimFlagDropDelimInReturn,
		buf)
	var tokens []string
	for s.Scan() {
		tokens = append(tokens, s.Text())
	}
	assert.NoError(t, s.Err())
	assert.Equal(t, []string{"abc#123", "efg", "", "???##xyz"}, tokens)
}

// Benchmark shows the benefit of using NewScannerByDelim3 with pre-allocated buf.
// BenchmarkNewScannerByDelim2-8                            	    5000	    299188 ns/op	 2141712 B/op	     996 allocs/op
// BenchmarkNewScannerByDelim3-8                            	   30000	     48893 ns/op	     208 B/op	       3 allocs/op

var benchmarkInput = strings.Repeat("abc#", 1000)

func BenchmarkNewScannerByDelim2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := NewScannerByDelim2(
			strings.NewReader(benchmarkInput),
			"#",
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

func BenchmarkNewScannerByDelim3(b *testing.B) {
	buf := make([]byte, 0, 10)
	for i := 0; i < b.N; i++ {
		s := NewScannerByDelim3(
			strings.NewReader(benchmarkInput),
			"#",
			nil,
			ScannerByDelimFlagEofAsDelim|ScannerByDelimFlagDropDelimInReturn,
			buf)
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
