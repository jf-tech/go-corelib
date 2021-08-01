package ios

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineNumReportingCsvReader(t *testing.T) {
	r := NewLineNumReportingCsvReader(strings.NewReader("a,b,c"))
	assert.Equal(t, 0, r.LineNum())
	record, err := r.Read()
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, record)
	assert.Equal(t, 1, r.LineNum())
	record, err = r.Read()
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, record)
	assert.Equal(t, 2, r.LineNum())

	r = NewLineNumReportingCsvReader(strings.NewReader("a,b,c"))
	r.numLineFieldName = "non-existing"
	assert.PanicsWithValue(
		t,
		"unable to get 'non-existing' from csv.Reader, has csv.Reader been changed/upgraded?",
		func() {
			r.LineNum()
		})
}

func TestByteReadLineAndReadLine(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		bufsize  int
		expected [][]byte
	}{
		{
			name:     "empty",
			input:    "",
			bufsize:  1024,
			expected: nil,
		},
		{
			name:    "single-line with no newline",
			input:   "   word1, word2 - word3 !@#$%^&*()",
			bufsize: 1024,
			expected: [][]byte{
				[]byte("   word1, word2 - word3 !@#$%^&*()"),
			},
		},
		{
			name:     "single-line with CR and LF",
			input:    "line1\r\n",
			bufsize:  1024,
			expected: [][]byte{[]byte("line1")},
		},
		{
			name:     "multi-line - bufsize enough",
			input:    "line1\r\nline2\nline3",
			bufsize:  1024,
			expected: [][]byte{[]byte("line1"), []byte("line2"), []byte("line3")},
		},
		{
			name:     "multi-line - bufsize not enough; also empty line",
			input:    "line1-0123456789012345\r\n\nline3-0123456789012345",
			bufsize:  16, // bufio.minReadBufferSize is 16.
			expected: [][]byte{[]byte("line1-0123456789012345"), nil, []byte("line3-0123456789012345")},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := bufio.NewReaderSize(strings.NewReader(test.input), test.bufsize)
			var outputBytes [][]byte
			for {
				line, err := ByteReadLine(r)
				if err != nil {
					assert.Nil(t, line)
					assert.Equal(t, io.EOF, err)
					break
				}
				if line != nil {
					// note the []byte returned by ByteReadLine can be invalidated upon next call, so
					// let's make a copy of it, thus the seemingly unnecessary `[]byte(string(line))`.
					outputBytes = append(outputBytes, []byte(string(line)))
				} else {
					outputBytes = append(outputBytes, line)
				}
			}
			assert.Equal(t, test.expected, outputBytes)

			r = bufio.NewReaderSize(strings.NewReader(test.input), test.bufsize)
			outputBytes = outputBytes[:0]
			for {
				line, err := ReadLine(r)
				if err != nil {
					assert.Equal(t, "", line)
					assert.Equal(t, io.EOF, err)
					break
				}
				if line == "" {
					outputBytes = append(outputBytes, nil)
				} else {
					outputBytes = append(outputBytes, []byte(line))
				}
			}
			assert.Equal(t, test.expected, outputBytes)
		})
	}
}

func TestStripBOM_Success(t *testing.T) {
	for _, test := range []struct {
		name            string
		fileContent     []byte
		expectedContent []byte
	}{
		{
			name:            "Empty content",
			fileContent:     []byte(""),
			expectedContent: []byte(""),
		},
		{
			name:            "Non-bom unicode",
			fileContent:     []byte("\u1234test content"),
			expectedContent: []byte("\u1234test content"),
		},
		{
			name:            "Content without BOM",
			fileContent:     []byte("test content"),
			expectedContent: []byte("test content"),
		},
		{
			name:            "Content with BOM",
			fileContent:     []byte("\uFEFFtest content"),
			expectedContent: []byte("test content"),
		},
		{
			name:            "Content with BOM only",
			fileContent:     []byte("\uFEFF"),
			expectedContent: []byte(""),
		},
	} {
		r := bytes.NewReader(test.fileContent)
		br, err := StripBOM(r)
		assert.NoError(t, err)
		assert.False(t, br == nil)
		line, _, err := br.(*bufio.Reader).ReadLine()
		if len(test.expectedContent) <= 0 {
			assert.Error(t, err)
			assert.Equal(t, io.EOF, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expectedContent, line)
	}
}

type failureReader struct {
	bytesToReturn []byte
	err           string
}

func (r *failureReader) Read(p []byte) (int, error) {
	length := len(r.bytesToReturn)
	if length > len(p) {
		length = len(p)
	}
	if length > 0 {
		copy(p, r.bytesToReturn)
	}
	if r.err != "" {
		return length, errors.New(r.err)
	}
	return length, errors.New("test failure")
}

func TestStripBOM_ReadFailure(t *testing.T) {
	br, err := StripBOM(&failureReader{})
	assert.True(t, br == nil)
	assert.Error(t, err)
	assert.Equal(t, "test failure", err.Error())
}

func TestLineCountingReader(t *testing.T) {
	for _, test := range []struct {
		name           string
		input          io.Reader
		expectedErr    string
		expectedBytes  []byte
		expectedAtLine int
	}{
		{
			name:           "success - one line",
			input:          strings.NewReader("abc efg"),
			expectedErr:    "",
			expectedBytes:  []byte("abc efg"),
			expectedAtLine: 1,
		},
		{
			name:           "success - multiple lines",
			input:          strings.NewReader("abc\nefg\r\n123\n"),
			expectedErr:    "",
			expectedBytes:  []byte("abc\nefg\r\n123\n"),
			expectedAtLine: 4,
		},
		{
			name:           "error - lines counted",
			input:          &failureReader{bytesToReturn: []byte("\n\n")},
			expectedErr:    "test failure",
			expectedBytes:  []byte("\n\n"),
			expectedAtLine: 3,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := NewLineCountingReader(test.input)
			b, err := ioutil.ReadAll(r)
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedBytes, b)
			assert.Equal(t, test.expectedAtLine, r.AtLine())
		})
	}
}
