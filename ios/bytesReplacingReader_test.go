package ios

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tjarratt/babble"
	"io/ioutil"
	"math/rand"
	"strings"
	"testing"
)

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

func createMultiAlphanumericalTestInput(length int, numTokens int, tokenLength int) ([]byte, [][]byte, [][]byte) {
	rand.Seed(1234)
	b := make([]byte, length)
	babbler := babble.NewBabbler()
	words := make([][]byte, 64)
	for i := 0; i < 64; i++ {
		word := babbler.Babble()
		words[i] = []byte(word)
	}
	tokenWords := make([][]byte, numTokens)
	for i := 0; i < numTokens; i++ {
		tokenWords[i] = words[rand.Intn(63)]
	}
	for i := 0; i < length; {
		r := make([]byte, 0)
		for cur := 0; cur < tokenLength; {
			w := words[rand.Intn(63)]
			r = append(r, w...)
			cur += len(w)
		}
		b = append(b, r...)
		i += len(r)
	}
	replaces := make([][]byte, numTokens)
	for i := 0; i < numTokens; i++ {
		replaces[i] = []byte(fmt.Sprintf("REPLACED-%d", i))
	}
	return b, tokenWords, replaces
}

var testInput70MBLength500Targets = createTestInput(100*1024*1024, 500)
var testInput1000MBLength64Targets, testInput1000MBLength64TargetsTokens, testInput1000MBLength64TargetsReplaces = createMultiAlphanumericalTestInput(1000*1024*1024, 1024, 1024)
var testInput1KBLength20Targets = createTestInput(1024, 20)
var testInput50KBLength1000Targets = createTestInput(50*1024, 1000)
var testSearchFor = []byte{7}
var testReplaceWith = []byte{8}
var testReplacer = &singleSearchReplaceReplacer{search: testSearchFor, replace: testReplaceWith}
var testMultiTokenReplacer = &multiTokenReplacer{searches: testInput1000MBLength64TargetsTokens, replaces: testInput1000MBLength64TargetsReplaces}
func BenchmarkBytesReplacingReader_70MBLength_500Targets(b *testing.B) {
	r := &BytesReplacingReader{}
	for i := 0; i < b.N; i++ {
		r.ResetEx(bytes.NewReader(testInput70MBLength500Targets), testReplacer)
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
		r.ResetEx(bytes.NewReader(testInput1KBLength20Targets), testReplacer)
		_, _ = ioutil.ReadAll(r)
	}
}

func BenchmarkBytesMultiTokenReader_100MBLength_64Targets(b *testing.B) {
	r := &BytesReplacingReader{}
	for i := 0; i < b.N; i++ {
		r.ResetEx(bytes.NewReader(testInput1000MBLength64Targets), testMultiTokenReplacer)
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
		r.ResetEx(bytes.NewReader(testInput50KBLength1000Targets), testReplacer)
		_, _ = ioutil.ReadAll(r)
	}
}

func BenchmarkRegularReader_50KBLength_1000Targets(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ioutil.ReadAll(bytes.NewReader(testInput50KBLength1000Targets))
	}
}

// The follow struct/test is to demonstrate how to do a different customization of BytesReplacer.
type multiTokenReplacer struct {
	searches [][]byte
	replaces [][]byte
}

func (r *multiTokenReplacer) GetSizingHints() (int, int, float64) {
	if len(r.searches) != len(r.replaces) {
		panic(fmt.Sprintf("len(searches) (%d) != len(replaces) (%d)", len(r.searches), len(r.replaces)))
	}
	if len(r.searches) == 0 {
		panic("searches must have at least one token")
	}
	maxSearchLen := 0
	maxReplaceLen := 0
	maxRatio := float64(-1)
	for i := range r.searches {
		searchLen := len(r.searches[i])
		replaceLen := len(r.replaces[i])
		if searchLen > maxSearchLen {
			maxSearchLen = searchLen
		}
		if replaceLen > maxReplaceLen {
			maxReplaceLen = replaceLen
		}
		if searchLen < replaceLen {
			ratio := float64(searchLen) / float64(replaceLen)
			if ratio > maxRatio {
				maxRatio = ratio
			}
		}
	}
	return maxSearchLen, maxReplaceLen, maxRatio
}

func (r *multiTokenReplacer) BestIndex(buf []byte) (int, []byte, []byte) {
	for i := range r.searches {
		index := bytes.Index(buf, r.searches[i])
		if index >= 0 {
			return index, r.searches[i], r.replaces[i]
		}
	}
	return -1, nil, nil
}

func TestMultiTokenBytesReplacingReader(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    []byte
		searches [][]byte
		replaces [][]byte
		expected []byte
	}{
		{
			name:  "multi tokens; len(search) < len(replace); len(search) > len(replace); replace = nil",
			input: []byte("abcdefgop01234qrstuvwxyz"),
			searches: [][]byte{
				[]byte("abc"),
				[]byte("12"),
				[]byte("st"),
				[]byte("xyz"),
			},
			replaces: [][]byte{
				[]byte("one two three"),
				[]byte("twelve is an int"),
				nil,
				[]byte("uv"),
			},
			expected: []byte("one two threedefgop0twelve is an int34qruvwuv"),
		},
	} {
		replacer := &multiTokenReplacer{
			searches: test.searches,
			replaces: test.replaces,
		}
		r := NewBytesReplacingReaderEx(bytes.NewReader(test.input), replacer)
		result, err := ioutil.ReadAll(r)
		assert.NoError(t, err)
		assert.Equal(t, string(test.expected), string(result))
	}

	r := (&BytesReplacingReader{}).ResetEx(
		strings.NewReader("test"),
		&multiTokenReplacer{
			searches: [][]byte{[]byte("abc"), []byte("")},
			replaces: [][]byte{[]byte("xyz"), []byte("wrong")},
		})
	assert.PanicsWithValue(t, "search token cannot be nil/empty", func() {
		_, _ = ioutil.ReadAll(r)
	})
}
