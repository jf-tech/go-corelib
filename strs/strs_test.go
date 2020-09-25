package strs

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestIndexWithEsc(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		delim    string
		esc      *rune
		expected int
	}{
		// All edge cases:
		{
			name:     "delim empty",
			input:    "abc",
			delim:    "",
			esc:      RunePtr(rune('宇')),
			expected: 0,
		},
		{
			name:     "esc empty",
			input:    "abc",
			delim:    "bc",
			esc:      nil,
			expected: 1,
		},
		{
			name:     "input empty, delim non empty, esc non empty",
			input:    "",
			delim:    "abc",
			esc:      RunePtr(rune('宙')),
			expected: -1,
		},
		// normal non empty cases:
		{
			name:     "len(input) < len(delim)",
			input:    "a",
			delim:    "abc",
			esc:      RunePtr(rune('洪')),
			expected: -1,
		},
		{
			name:     "len(input) == len(delim), esc not present",
			input:    "abc",
			delim:    "abc",
			esc:      RunePtr(rune('荒')),
			expected: 0,
		},
		{
			name:     "len(input) > len(delim), esc not present",
			input:    "мир во всем мире",
			delim:    "мире",
			esc:      RunePtr(rune('Ф')),
			expected: len("мир во всем "),
		},
		{
			name:     "len(input) > len(delim), esc present",
			input:    "мир во всем /мире",
			delim:    "мире",
			esc:      RunePtr(rune('/')),
			expected: -1,
		},
		{
			name:     "len(input) > len(delim), esc present",
			input:    "мир во всем ξξмире",
			delim:    "мире",
			esc:      RunePtr(rune('ξ')),
			expected: len("мир во всем ξξ"),
		},
		{
			name:     "len(input) > len(delim), consecutive esc present",
			input:    "мир во вξξξξξсем ξξмире",
			delim:    "ире",
			esc:      RunePtr(rune('ξ')),
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

func TestSplitWithEsc(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		delim    string
		esc      *rune
		expected []string
	}{
		{
			name:     "delim empty",
			input:    "abc",
			delim:    "",
			esc:      RunePtr(rune('宇')),
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "esc not set",
			input:    "",
			delim:    "abc",
			esc:      nil,
			expected: []string{""},
		},
		{
			name:     "esc set, delim not found",
			input:    "?xyz",
			delim:    "xyz",
			esc:      RunePtr(rune('?')),
			expected: []string{"?xyz"},
		},
		{
			name:     "esc set, delim found",
			input:    "a*bc/*d*efg",
			delim:    "*",
			esc:      RunePtr(rune('/')),
			expected: []string{"a", "bc/*d", "efg"},
		},
		{
			name:     "esc set, delim not empty, input empty",
			input:    "",
			delim:    "*",
			esc:      RunePtr(rune('/')),
			expected: []string{""},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, SplitWithEsc(test.input, test.delim, test.esc))
		})
	}
}

func TestUnescape(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		esc      *rune
		expected string
	}{
		{
			name:     "esc not set",
			input:    "abc",
			esc:      nil,
			expected: "abc",
		},
		{
			name:     "esc set, input empty",
			input:    "",
			esc:      RunePtr(rune('宇')),
			expected: "",
		},
		{
			name:     "esc set, input non empty",
			input:    "ξξabcξdξ",
			esc:      RunePtr(rune('ξ')),
			expected: "ξabcd",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, Unescape(test.input, test.esc))
		})
	}
}
