package iohelper

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
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

func TestBytesReplacingReader(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    []byte
		search   []byte
		replace  []byte
		expected []byte
	}{
		{
			name:     "len(replace) > len(search)",
			input:    []byte{1, 2, 3, 2, 2, 3, 4, 5},
			search:   []byte{2, 3},
			replace:  []byte{4, 5, 6},
			expected: []byte{1, 4, 5, 6, 2, 4, 5, 6, 4, 5},
		},
		{
			name:     "len(replace) < len(search)",
			input:    []byte{1, 2, 3, 2, 2, 3, 4, 5, 6, 7, 8},
			search:   []byte{2, 3, 2},
			replace:  []byte{9},
			expected: []byte{1, 9, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name:     "strip out search, no replace",
			input:    []byte{1, 2, 3, 2, 2, 3, 4, 2, 3, 2, 8},
			search:   []byte{2, 3, 2},
			replace:  []byte{},
			expected: []byte{1, 2, 3, 4, 8},
		},
		{
			name:     "len(replace) == len(search)",
			input:    []byte{1, 2, 3, 4, 5, 5, 5, 5, 5, 5, 5, 5, 5},
			search:   []byte{5, 5},
			replace:  []byte{6, 6},
			expected: []byte{1, 2, 3, 4, 6, 6, 6, 6, 6, 6, 6, 6, 5},
		},
		{
			name:     "double quote -> single quote",
			input:    []byte(`r = NewLineNumReportingCsvReader(strings.NewReader("a,b,c"))`),
			search:   []byte(`"`),
			replace:  []byte(`'`),
			expected: []byte(`r = NewLineNumReportingCsvReader(strings.NewReader('a,b,c'))`),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := NewBytesReplacingReader(bytes.NewReader(test.input), test.search, test.replace)
			result, err := ioutil.ReadAll(r)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)

		})
	}

	assert.PanicsWithValue(t, "io.Reader cannot be nil", func() {
		NewBytesReplacingReader(nil, []byte{1}, []byte{2})
	})
	assert.PanicsWithValue(t, "search token cannot be nil/empty", func() {
		(&BytesReplacingReader{}).Reset(strings.NewReader("test"), nil, []byte("est"))
	})
}

func createTestInput(length int, numTarget int) []byte {
	rand.Seed(1234) // fixed rand seed to ensure bench stability
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = byte(rand.Intn(100) + 10) // all regular numbers >= 10
	}
	for i := 0; i < numTarget; i++ {
		for {
			index := rand.Intn(length)
			if b[index] == 7 {
				continue
			}
			b[index] = 7 // special number 7 we will search for and replace with 8.
			break
		}
	}
	return b
}

var testInput70MBLength500Targets = createTestInput(70*1024*1024, 500)
var testInput1KBLength20Targets = createTestInput(1024, 20)
var testInput50KBLength1000Targets = createTestInput(50*1024, 1000)
var testSearchFor = []byte{7}
var testReplaceWith = []byte{8}

func BenchmarkBytesReplacingReader_70MBLength_500Targets(b *testing.B) {
	r := &BytesReplacingReader{}
	for i := 0; i < b.N; i++ {
		r.Reset(bytes.NewReader(testInput70MBLength500Targets), testSearchFor, testReplaceWith)
		_, _ = ioutil.ReadAll(r)
	}
}

func BenchmarkRegularReader_70MBLength_500Targets(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ioutil.ReadAll(bytes.NewReader(testInput70MBLength500Targets))
	}
}

func BenchmarkBytesReplacingReader_1KBLength_20Targets(b *testing.B) {
	r := &BytesReplacingReader{}
	for i := 0; i < b.N; i++ {
		r.Reset(bytes.NewReader(testInput1KBLength20Targets), testSearchFor, testReplaceWith)
		_, _ = ioutil.ReadAll(r)
	}
}

func BenchmarkRegularReader_1KBLength_20Targets(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ioutil.ReadAll(bytes.NewReader(testInput1KBLength20Targets))
	}
}

func BenchmarkBytesReplacingReader_50KBLength_1000Targets(b *testing.B) {
	r := &BytesReplacingReader{}
	for i := 0; i < b.N; i++ {
		r.Reset(bytes.NewReader(testInput50KBLength1000Targets), testSearchFor, testReplaceWith)
		_, _ = ioutil.ReadAll(r)
	}
}

func BenchmarkRegularReader_50KBLength_1000Targets(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ioutil.ReadAll(bytes.NewReader(testInput50KBLength1000Targets))
	}
}

func TestReadLine(t *testing.T) {
	for _, test := range []struct {
		name           string
		input          string
		bufsize        int
		expectedOutput []string
	}{
		{
			name:           "empty",
			input:          "",
			bufsize:        1024,
			expectedOutput: []string{},
		},
		{
			name:           "single-line with no newline",
			input:          "   word1, word2 - word3 !@#$%^&*()",
			bufsize:        1024,
			expectedOutput: []string{"   word1, word2 - word3 !@#$%^&*()"},
		},
		{
			name:           "single-line with '\\r' and '\\n'",
			input:          "line1\r\n",
			bufsize:        1024,
			expectedOutput: []string{"line1"},
		},
		{
			name:           "multi-line - bufsize enough",
			input:          "line1\r\nline2\nline3",
			bufsize:        1024,
			expectedOutput: []string{"line1", "line2", "line3"},
		},
		{
			name:           "multi-line - bufsize not enough; also empty line",
			input:          "line1-0123456789012345\r\n\nline3-0123456789012345",
			bufsize:        16, // bufio.minReadBufferSize is 16.
			expectedOutput: []string{"line1-0123456789012345", "", "line3-0123456789012345"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := bufio.NewReaderSize(strings.NewReader(test.input), test.bufsize)
			output := []string{}
			for {
				line, err := ReadLine(r)
				if err != nil {
					assert.Equal(t, "", line)
					assert.Equal(t, io.EOF, err)
					break
				}
				output = append(output, line)
			}
			assert.Equal(t, test.expectedOutput, output)
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
