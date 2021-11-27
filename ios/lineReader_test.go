package ios

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineReaderWithCustomBufSize(t *testing.T) {
	for _, test := range []struct {
		name     string
		editFunc LineEditFunc
		bufSize  int
		input    string
		expected string
		err      string
	}{
		{
			name: "various successful editings",
			editFunc: func(line []byte) ([]byte, error) {
				if string(line) == "abc\n" {
					// testing returning a newly allocated line with same length
					return []byte("xyz\n"), nil
				}
				if string(line) == "one\r\n" {
					line[0] = '1'
					line[1] = '\r'
					line[2] = '\n'
					// testing an in-place edited line with shrunk length
					return line[:3], nil
				}
				if string(line) == "1" { // note there is no ending '\n' since line "1" is the last line before EOF.
					// testing returning a newly allocated line with much longer length plus some '\n' added.
					return []byte("first\nzuerst\nprimo\n第一の"), nil
				}
				return line, nil
			},
			bufSize:  2,
			input:    "not changed\nabc\n\n\n\none\r\n1",
			expected: "not changed\nxyz\n\n\n\n1\r\nfirst\nzuerst\nprimo\n第一の",
			err:      "",
		},
		{
			name: "successful editing followed by failed editing",
			editFunc: func(line []byte) ([]byte, error) {
				if string(line) == "abc\n" {
					return []byte("xyz\n"), nil
				}
				if string(line) == "boom\r\n" {
					return []byte("ignored\r\n"), errors.New("mock error")
				}
				return line, nil
			},
			bufSize:  100,
			input:    "not changed\nabc\nboom\r\nend\n",
			expected: "",
			err:      "mock error",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ret, err := ioutil.ReadAll(NewLineReader2(strings.NewReader(test.input), test.editFunc, test.bufSize))
			if test.err != "" {
				assert.Error(t, err)
				assert.Equal(t, test.err, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, string(ret))
			}
		})
	}
}

func TestNewLineReader(t *testing.T) {
	// Test against a real scenario where we need to strip each line's leading '|' pipe character.
	// See details at: https://github.com/jf-tech/omniparser/pull/154
	input := "|HDR|1|2|3|\n\n|DAT|X|\n|EOF|"
	expected := "HDR|1|2|3|\n\nDAT|X|\nEOF|"
	ret, err := ioutil.ReadAll(
		NewLineReader(
			strings.NewReader(input),
			func(line []byte) ([]byte, error) {
				if len(line) < 2 || line[0] != '|' {
					return line, nil
				}
				return line[1:], nil
			}))
	assert.NoError(t, err)
	assert.Equal(t, expected, string(ret))
}

var (
	lineReaderBenchInputLine = "|HDR|1|2|3|4|5|6|7|8|9|\n"
	lineReaderBenchInput     = strings.Repeat(lineReaderBenchInputLine, 10000)
	lineReaderBenchOutput    = strings.Repeat(strings.TrimLeft(lineReaderBenchInputLine, "|"), 10000)
)

func TestLineReaderBench(t *testing.T) {
	ret, err := ioutil.ReadAll(
		NewLineReader(
			strings.NewReader(lineReaderBenchInput),
			func(line []byte) ([]byte, error) {
				if len(line) < 2 || line[0] != '|' {
					return line, nil
				}
				return line[1:], nil
			}))
	assert.NoError(t, err)
	assert.Equal(t, lineReaderBenchOutput, string(ret))
}

func BenchmarkLineReader_RawIORead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ioutil.ReadAll(strings.NewReader(lineReaderBenchInput))
	}
}

func BenchmarkLineReader_UseLineReader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ioutil.ReadAll(
			NewLineReader(
				strings.NewReader(lineReaderBenchInput),
				func(line []byte) ([]byte, error) {
					if len(line) < 2 || line[0] != '|' {
						return line, nil
					}
					return line[1:], nil
				}))
	}
}

func BenchmarkLineReader_CompareWithBytesReplacingReader(b *testing.B) {
	search := []byte("|H")
	replace := []byte("H")
	for i := 0; i < b.N; i++ {
		_, _ = ioutil.ReadAll(
			NewBytesReplacingReader(
				strings.NewReader(lineReaderBenchInput),
				search,
				replace))
	}
}
