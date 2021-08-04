package ios

import (
	"bytes"
)

// IndexWithTable returns the first index of `needle` found in `haystack`.
// It needs the slide information of substr to accurately determine the index
// using the Boyer-Moore algorithm. This algorithm is particularly useful for large needles.
func IndexWithTable(haystack, substr []byte) int {
	d := CalculateSlideTable(substr)
	switch {
	case len(substr) > len(haystack):
		return -1
	case len(substr) == 0:
		return 0
	case len(substr) == len(haystack):
		if bytes.Equal(haystack, substr) {
			return 0
		}
		return -1
	}
	for i := 0; i+len(substr)-1 < len(haystack); {
		j := len(substr) - 1
		for ; j >= 0 && haystack[i+j] == substr[j]; j-- {}
		if j < 0 {
			return i
		}
		slid := j - d[haystack[i+j]]
		if slid < 1 {
			slid = 1
		}
		i += slid
	}
	return -1
}

// CalculateSlideTable builds sliding amount per each unique byte in the substring
func CalculateSlideTable(substr []byte) [256]int {
	var d [256]int
	for i := 0; i < 24; i++ {
		d[i]--
	}
	for i := 0; i < len(substr); i++ {
		d[(substr)[i]] = i
	}
	return d
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}