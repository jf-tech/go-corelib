package strs

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunePtr(t *testing.T) {
	rp := RunePtr('p')
	assert.NotNil(t, rp)
	assert.Equal(t, 'p', *rp)
}

func TestStrPtr(t *testing.T) {
	sp := StrPtr("pi")
	assert.NotNil(t, sp)
	assert.Equal(t, "pi", *sp)
}

func TestIsStrNonBlank(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		nonBlank bool
	}{
		{
			name:     "empty string",
			input:    "",
			nonBlank: false,
		},
		{
			name:     "blank string",
			input:    "      ",
			nonBlank: false,
		},
		{
			name:     "non blank",
			input:    "abc",
			nonBlank: true,
		},
		{
			name:     "non blank after trimming",
			input:    "  abc  ",
			nonBlank: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.nonBlank, IsStrNonBlank(test.input))
			inputCopy := test.input
			assert.Equal(t, test.nonBlank, IsStrPtrNonBlank(&inputCopy))
		})
	}
	assert.False(t, IsStrPtrNonBlank(nil))
}

func TestFirstNonBlank(t *testing.T) {
	assert.Equal(t, "abc", FirstNonBlank("", "   ", "abc", "def"))
	assert.Equal(t, "", FirstNonBlank("", "   ", "           "))
	assert.Equal(t, "", FirstNonBlank())
}

func TestStrPtrOrElse(t *testing.T) {
	assert.Equal(t, "this", StrPtrOrElse(StrPtr("this"), "that"))
	assert.Equal(t, "that", StrPtrOrElse(nil, "that"))
}

func TestCopyStrPtr(t *testing.T) {
	assert.True(t, CopyStrPtr(nil) == nil)
	src := StrPtr("abc")
	dst := CopyStrPtr(src)
	assert.Equal(t, *src, *dst)
	assert.True(t, fmt.Sprintf("%p", src) != fmt.Sprintf("%p", dst))
}

func TestBuildFQDN(t *testing.T) {
	for _, test := range []struct {
		name     string
		namelets []string
		expected string
	}{
		{
			name:     "nil",
			namelets: nil,
			expected: "",
		},
		{
			name:     "empty",
			namelets: []string{},
			expected: "",
		},
		{
			name:     "single",
			namelets: []string{"one"},
			expected: "one",
		},
		{
			name:     "multiple",
			namelets: []string{"one", "", "three", "four"},
			expected: "one..three.four",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, BuildFQDN(test.namelets...))
		})
	}
}

func TestLastNameletOfFQDN(t *testing.T) {
	for _, test := range []struct {
		name     string
		fqdn     string
		expected string
	}{
		{
			name:     "empty",
			fqdn:     "",
			expected: "",
		},
		{
			name:     "no delimiter",
			fqdn:     "abc",
			expected: "abc",
		},
		{
			name:     "delimiter at beginning",
			fqdn:     ".abc",
			expected: "abc",
		},
		{
			name:     "delimiter at the end",
			fqdn:     "abc.",
			expected: "",
		},
		{
			name:     "fqdn",
			fqdn:     "abc.def.ghi",
			expected: "ghi",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, LastNameletOfFQDN(test.fqdn))
		})
	}
}

func TestCopySlice(t *testing.T) {
	for _, test := range []struct {
		name           string
		input          []string
		expectedOutput []string
	}{
		{
			name:           "nil",
			input:          nil,
			expectedOutput: nil,
		},
		{
			name:           "empty slice",
			input:          []string{},
			expectedOutput: nil,
		},
		{
			name:           "non-empty slice",
			input:          []string{"abc", ""},
			expectedOutput: []string{"abc", ""},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			cp := CopySlice(test.input)
			// First make sure the copy contains what's expected.
			assert.Equal(t, test.expectedOutput, cp)
			if len(test.input) >= 2 {
				// Second test if modifying the original won't affect the copy
				// (that's what this copy func is all about)
				test.input[0] = test.input[1]
				assert.NotEqual(t, test.input, cp)
			}
		})
	}
}

func TestMergeSlices(t *testing.T) {
	for _, test := range []struct {
		name     string
		slice1   []string
		slice2   []string
		expected []string
	}{
		{
			name:     "both nil",
			slice1:   nil,
			slice2:   nil,
			expected: nil,
		},
		{
			name:     "1 nil, 2 not nil",
			slice1:   nil,
			slice2:   []string{"", "abc"},
			expected: []string{"", "abc"},
		},
		{
			name:     "1 not nil, 2 nil",
			slice1:   []string{"abc", ""},
			slice2:   nil,
			expected: []string{"abc", ""},
		},
		{
			name:     "both not nil",
			slice1:   []string{"abc", ""},
			slice2:   []string{"", "abc"},
			expected: []string{"abc", "", "", "abc"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			merged := MergeSlices(test.slice1, test.slice2)
			// also very importantly to make sure the resulting merged is a new copy so modifying
			// the input slices won't affect the merged slice.
			if len(test.slice1) > 0 {
				test.slice1[0] = "modified"
			}
			if len(test.slice2) > 0 {
				test.slice2[0] = "modified"
			}
			assert.Equal(t, test.expected, merged)
		})
	}
}

func TestHasDup(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    []string
		expected bool
	}{
		{
			name:     "nil",
			input:    nil,
			expected: false,
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: false,
		},
		{
			name:     "non-empty slice with no dups",
			input:    []string{"abc", ""},
			expected: false,
		},
		{
			name:     "non-empty slice with dups",
			input:    []string{"", "abc", ""},
			expected: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, HasDup(test.input))
		})
	}
}

func TestMapSlice(t *testing.T) {
	t.Run("map error", func(t *testing.T) {
		errorMap := func(_ string) (string, error) {
			return "abc", errors.New("map error")
		}
		result, err := MapSlice([]string{"abc", ""}, errorMap)
		assert.Error(t, err)
		assert.Equal(t, "map error", err.Error())
		assert.Nil(t, result)
	})

	t.Run("map success", func(t *testing.T) {
		input := []string{"abc", ""}
		index := 0
		mirrorMap := func(_ string) (string, error) {
			index++
			return input[len(input)-index], nil
		}
		result, err := MapSlice(input, mirrorMap)
		assert.NoError(t, err)
		assert.Equal(t, []string{"", "abc"}, result)
	})

	t.Run("map nil", func(t *testing.T) {
		result, err := MapSlice(nil, func(s string) (string, error) {
			return s + "...", nil
		})
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestNoErrMapSlice(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "non-empty slice",
			input:    []string{"abc", ""},
			expected: []string{"", "abc"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			index := 0
			mirrorMap := func(s string) string {
				index++
				return test.input[len(test.input)-index]
			}
			assert.Equal(t, test.expected, NoErrMapSlice(test.input, mirrorMap))
		})
	}
}

func TestByteLenOfRunes(t *testing.T) {
	assert.Equal(t, 0, ByteLenOfRunes(nil))
	assert.Equal(t, 0, ByteLenOfRunes([]rune{}))
	assert.Equal(t, 10, ByteLenOfRunes([]rune("0123456789")))
	assert.Equal(t, 30, ByteLenOfRunes([]rune("壹贰叁肆伍陆柒捌玖拾")))
}

func TestIndexWithEsc(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		delim    string
		esc      string
		expected int
	}{
		// All edge cases:
		{
			name:     "delim empty",
			input:    "abc",
			delim:    "",
			esc:      "宇",
			expected: 0,
		},
		{
			name:     "esc empty",
			input:    "abc",
			delim:    "bc",
			esc:      "",
			expected: 1,
		},
		{
			name:     "input empty, delim non empty, esc non empty",
			input:    "",
			delim:    "abc",
			esc:      "宙",
			expected: -1,
		},
		// normal non empty cases:
		{
			name:     "len(input) < len(delim)",
			input:    "a",
			delim:    "abc",
			esc:      "洪",
			expected: -1,
		},
		{
			name:     "len(input) == len(delim), esc not present",
			input:    "abc",
			delim:    "abc",
			esc:      "荒",
			expected: 0,
		},
		{
			name:     "len(input) > len(delim), esc not present",
			input:    "мир во всем мире",
			delim:    "мире",
			esc:      "Ф",
			expected: len("мир во всем "),
		},
		{
			name:     "len(input) > len(delim), esc present",
			input:    "мир во всем /мире м/ире",
			delim:    "мире",
			esc:      "/",
			expected: -1,
		},
		{
			name:     "len(input) > len(delim), esc present",
			input:    "мир во всем ξξмире",
			delim:    "мире",
			esc:      "ξ",
			expected: len("мир во всем ξξ"),
		},
		{
			name:     "len(input) > len(delim), consecutive esc present",
			input:    "мир во вξξξξξсем ξξмире",
			delim:    "ире",
			esc:      "ξ",
			expected: len("мир во вξξξξξсем ξξм"),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, IndexWithEsc(test.input, test.delim, test.esc))
			if test.expected >= 0 {
				assert.True(t, strings.HasPrefix(string([]byte(test.input)[test.expected:]), test.delim))
			}
		})
	}
}

// BenchmarkIndexWithEsc-8       	50000000	        28.9 ns/op	       0 B/op	       0 allocs/op
func BenchmarkIndexWithEsc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if IndexWithEsc("мир во вξξξξξсем ξξмире", "ире", "ξ") < 0 {
			b.FailNow()
		}
	}
}

func TestSplitWithEsc(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		delim    string
		esc      string
		expected []string
	}{
		{
			name:     "delim empty",
			input:    "abc",
			delim:    "",
			esc:      "宇",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "esc not set",
			input:    "",
			delim:    "abc",
			esc:      "",
			expected: []string{""},
		},
		{
			name:     "esc set, delim not found",
			input:    "?xyz",
			delim:    "xyz",
			esc:      "?",
			expected: []string{"?xyz"},
		},
		{
			name:     "esc set, delim found",
			input:    "a*bc/*d*efg",
			delim:    "*",
			esc:      "/",
			expected: []string{"a", "bc/*d", "efg"},
		},
		{
			name:     "esc set, delim not empty, input empty",
			input:    "",
			delim:    "*",
			esc:      "/",
			expected: []string{""},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, SplitWithEsc(test.input, test.delim, test.esc))
		})
	}
}

func TestByteIndexWithEsc(t *testing.T) {
	for _, test := range []struct {
		name     string
		s        []byte
		delim    []byte
		esc      []byte
		expected int
	}{
		{
			name:     "esc nil; bytes.Index used",
			s:        []byte("abc#ef##g"),
			delim:    []byte("##"),
			esc:      nil,
			expected: 6,
		},
		{
			name:     "esc non-empty; delim nil; bytes.Index used",
			s:        []byte("abc#ef##g"),
			delim:    nil,
			esc:      []byte("%"),
			expected: 0,
		},
		{
			name:     "esc non-empty; no delim found",
			s:        []byte("abc%#ef%#"),
			delim:    []byte("##"),
			esc:      []byte("%"),
			expected: -1,
		},
		{
			name:     "esc non-empty; delim found, no esc involved",
			s:        []byte("abc#ef##g"),
			delim:    []byte("##"),
			esc:      []byte("%"),
			expected: 6,
		},
		{
			name:     "delim preceded by even number of esc",
			s:        []byte("ab%%%%##ef#g"),
			delim:    []byte("##"),
			esc:      []byte("%%"),
			expected: 6,
		},
		{
			name:     "delim preceded by odd number of esc",
			s:        []byte("%^%^%^###ef#g"),
			delim:    []byte("##"),
			esc:      []byte("%^"),
			expected: 7,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ByteIndexWithEsc(test.s, test.delim, test.esc))
		})
	}
}

func TestByteSplitWithEsc(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    []byte
		delim    []byte
		esc      []byte
		cap      int
		expected [][]byte
	}{
		{
			name:     "delim empty",
			input:    []byte("abc"),
			delim:    []byte{},
			esc:      []byte("宇"),
			cap:      0,
			expected: [][]byte{[]byte("a"), []byte("b"), []byte("c")},
		},
		{
			name:     "delim not found",
			input:    []byte("?xyz"),
			delim:    []byte("xyz"),
			esc:      []byte("?"),
			cap:      0,
			expected: [][]byte{[]byte("?xyz")},
		},
		{
			name:     "delim found",
			input:    []byte("a*bc/*d*efg"),
			delim:    []byte("*"),
			esc:      []byte("/"),
			cap:      1,
			expected: [][]byte{[]byte("a"), []byte("bc/*d"), []byte("efg")},
		},
		{
			name:     "delim not empty, input nil",
			input:    nil,
			delim:    []byte("*"),
			esc:      []byte("/"),
			cap:      0,
			expected: [][]byte{nil},
		},
		{
			name:     "delim not empty, input empty",
			input:    []byte{},
			delim:    []byte("*"),
			esc:      []byte("/"),
			cap:      0,
			expected: [][]byte{{}},
		},
		{
			name:     "multi-utf8-rune delim",
			input:    []byte("デΩリミタabcデリミタefデリミタgデリミタ"),
			delim:    []byte("デリミタ"),
			esc:      []byte("Ω"),
			cap:      10,
			expected: [][]byte{[]byte("デΩリミタabc"), []byte("ef"), []byte("g"), {}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ByteSplitWithEsc(test.input, test.delim, test.esc, test.cap))
		})
	}
}

const (
	splitBenchInput = "abc:efg:xyzadf%:dfa:afas:%:dfasdf:dfasdfwrqwer:3adfaz:dfa:d:d:::dfadsf:efq%"
	splitBenchDelim = ":"
	splitBenchEsc   = "%"
)

var (
	splitBenchInputBytes = []byte(splitBenchInput)
	splitBenchDelimBytes = []byte(splitBenchDelim)
	splitBenchEscBytes   = []byte(splitBenchEsc)

	splitBenchResult = []string{
		"abc",
		"efg",
		"xyzadf%:dfa",
		"afas",
		"%:dfasdf",
		"dfasdfwrqwer",
		"3adfaz",
		"dfa",
		"d",
		"d",
		"",
		"",
		"dfadsf",
		"efq%",
	}

	splitBenchResultBytes = func() [][]byte {
		var bb [][]byte
		for _, s := range splitBenchResult {
			bb = append(bb, []byte(s))
		}
		return bb
	}()
)

func TestForBenchmarkSplitWithEsc(t *testing.T) {
	assert.Equal(t, splitBenchResult, SplitWithEsc(splitBenchInput, splitBenchDelim, splitBenchEsc))
}

// BenchmarkSplitWithEsc-8       	 2000000	       848 ns/op	     496 B/op	       5 allocs/op
func BenchmarkSplitWithEsc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SplitWithEsc(splitBenchInput, splitBenchDelim, splitBenchEsc)
	}
}

func TestForBenchmarkByteSplitWithEsc(t *testing.T) {
	assert.Equal(t,
		splitBenchResultBytes,
		ByteSplitWithEsc(splitBenchInputBytes, splitBenchDelimBytes, splitBenchEscBytes, len(splitBenchResult)))
}

// BenchmarkByteSplitWithEsc-8   	 2000000	       637 ns/op	     352 B/op	       1 allocs/op
func BenchmarkByteSplitWithEsc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ByteSplitWithEsc(splitBenchInputBytes, splitBenchDelimBytes, splitBenchEscBytes, len(splitBenchResult))
	}
}

func TestUnescape(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		esc      string
		expected string
	}{
		{
			name:     "esc empty",
			input:    "abc",
			esc:      "",
			expected: "abc",
		},
		{
			name:     "ecs non-empty, input empty",
			input:    "",
			esc:      "宇",
			expected: "",
		},
		{
			name:     "esc non-empty, input non empty",
			input:    "ξξabcξdξξ",
			esc:      "ξξ",
			expected: "abcξd",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, Unescape(test.input, string(test.esc)))
		})
	}
}

func TestByteUnescape(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    []byte
		esc      []byte
		expected []byte
	}{
		{
			name:     "input nil",
			input:    nil,
			esc:      []byte("%"),
			expected: nil,
		},
		{
			name:     "input empty, esc nil",
			input:    []byte(""),
			esc:      nil,
			expected: []byte(""),
		},
		{
			name:     "input non empty, esc empty",
			input:    []byte("abc"),
			esc:      nil,
			expected: []byte("abc"),
		},
		{
			name:     "input non-empty, esc non empty, esc not found",
			input:    []byte("ξξabcξdξξ"),
			esc:      []byte("ξξ$"),
			expected: []byte("ξξabcξdξξ"),
		},
		{
			name:     "input non-empty, esc non empty, esc found",
			input:    []byte("ξξabcξdξξ"),
			esc:      []byte("ξξ"),
			expected: []byte("abcξd"),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ByteUnescape(test.input, test.esc))
		})
	}
}

var (
	benchmarkUnescapeInput       = strings.Repeat("abc#", 100000)
	benchmarkUnescapeInputBytes  = []byte(benchmarkUnescapeInput)
	benchmarkUnescapeDelim       = "#"
	benchmarkUnescapeDelimBytes  = []byte(benchmarkUnescapeDelim)
	benchmarkUnescapeResult      = strings.Repeat("abc", 100000)
	benchmarkUnescapeResultBytes = []byte(benchmarkUnescapeResult)
)

// BenchmarkUnescape-8           	    1000	   2321659 ns/op	  401810 B/op	       1 allocs/op
func BenchmarkUnescape(b *testing.B) {
	assert.Equal(b, benchmarkUnescapeResult, Unescape(benchmarkUnescapeInput, benchmarkUnescapeDelim))
	for i := 0; i < b.N; i++ {
		_ = Unescape(benchmarkUnescapeInput, benchmarkUnescapeDelim)
	}
}

// BenchmarkByteUnescape-8       	     500	   2482229 ns/op	  402211 B/op	       1 allocs/op
func BenchmarkByteUnescape(b *testing.B) {
	assert.Equal(b, benchmarkUnescapeResultBytes, ByteUnescape(benchmarkUnescapeInputBytes, benchmarkUnescapeDelimBytes))
	for i := 0; i < b.N; i++ {
		_ = ByteUnescape(benchmarkUnescapeInputBytes, benchmarkUnescapeDelimBytes)
	}
}
