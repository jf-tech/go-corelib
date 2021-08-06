package ios

import (
	"bytes"
	"io"
)

// BytesReplacer allows customization on how BytesReplacingReader does sizing estimate during
// initialization/reset and does search and replacement during the execution.
type BytesReplacer interface {
	// GetSizingHints returns hints for BytesReplacingReader to do sizing estimate and allocation.
	// Return values:
	// - 1st: max search token len
	// - 2nd: max replace token len
	// - 3rd: max (search_len / replace_len) ratio that is less than 1,
	//        if none of the search/replace ratio is less than 1, then return a negative number.
	// will only be called once during BytesReplacingReader initialization/reset.
	GetSizingHints() (int, int, float64)
	// BestIndex does token search for BytesReplacingReader.
	// Return values:
	// - 1st: index of the first found search token; -1, if not found;
	// - 2nd: the found search token; ignored if not found;
	// - 3rd: the matching replace token; ignored if not found;
	BestIndex(buf []byte) (int, []byte, []byte)
}

// BytesReplacingReader allows transparent replacement of a given token during read operation.
type BytesReplacingReader struct {
	replacer          BytesReplacer
	maxSearchTokenLen int
	r                 io.Reader
	err               error
	buf               []byte
	// buf[0:buf0]: bytes already processed; buf[buf0:buf1] bytes read in but not yet processed.
	buf0, buf1 int
	// because we need to replace 'search' with 'replace', this marks the max bytes we can read into buf
	max int
}

const defaultBufSize = 4096

// ResetEx allows reuse of a previous allocated `*BytesReplacingReader` for buf allocation optimization.
func (r *BytesReplacingReader) ResetEx(r1 io.Reader, replacer BytesReplacer) *BytesReplacingReader {
	if r1 == nil {
		panic("io.Reader cannot be nil")
	}
	r.replacer = replacer
	maxSearchTokenLen, maxReplaceTokenLen, maxSearchOverReplaceLenRatio := r.replacer.GetSizingHints()
	if maxSearchTokenLen == 0 {
		panic("search token cannot be nil/empty")
	}
	r.maxSearchTokenLen = maxSearchTokenLen
	r.r = r1
	r.err = nil
	bufSize := max(defaultBufSize, max(maxSearchTokenLen, maxReplaceTokenLen))
	if r.buf == nil || len(r.buf) < bufSize {
		r.buf = make([]byte, bufSize)
	}
	r.buf0 = 0
	r.buf1 = 0
	r.max = len(r.buf)
	if maxSearchOverReplaceLenRatio > 0 {
		// If len(search) < len(replace), then we have to assume the worst case:
		// what's the max bound value such that if we have consecutive 'search' filling up
		// the buf up to buf[:max], and all of them are placed with 'replace', and the final
		// result won't end up exceed the len(buf)?
		r.max = int(maxSearchOverReplaceLenRatio * float64(len(r.buf)))
	}
	return r
}

// Reset allows reuse of a previous allocated `*BytesReplacingReader` for buf allocation optimization.
// `search` cannot be nil/empty. `replace` can.
func (r *BytesReplacingReader) Reset(r1 io.Reader, search1, replace1 []byte) *BytesReplacingReader {
	return r.ResetEx(r1, &singleSearchReplaceReplacer{search: search1, replace: replace1})
}

// Read implements the `io.Reader` interface.
func (r *BytesReplacingReader) Read(p []byte) (int, error) {
	n := 0
	for {
		if r.buf0 > 0 {
			n = copy(p, r.buf[0:r.buf0])
			r.buf0 -= n
			r.buf1 -= n
			if r.buf1 == 0 && r.err != nil {
				return n, r.err
			}
			copy(r.buf, r.buf[n:r.buf1+n])
			return n, nil
		} else if r.err != nil {
			return 0, r.err
		}

		n, r.err = r.r.Read(r.buf[r.buf1:r.max])
		if n > 0 {
			r.buf1 += n
			for {
				index, search, replace := r.replacer.BestIndex(r.buf[r.buf0:r.buf1])
				if index < 0 {
					r.buf0 = max(r.buf0, r.buf1-r.maxSearchTokenLen+1)
					break
				}
				searchTokenLen := len(search)
				if searchTokenLen == 0 {
					panic("search token cannot be nil/empty")
				}
				replaceTokenLen := len(replace)
				lenDelta := replaceTokenLen - searchTokenLen
				index += r.buf0
				copy(r.buf[index+replaceTokenLen:r.buf1+lenDelta], r.buf[index+searchTokenLen:r.buf1])
				copy(r.buf[index:index+replaceTokenLen], replace)
				r.buf0 = index + replaceTokenLen
				r.buf1 += lenDelta
			}
		}
		if r.err != nil {
			r.buf0 = r.buf1
		}
	}
}

type singleSearchReplaceReplacer struct {
	search     []byte
	replace    []byte
	slide      [256]int
	searchLen  int
	replaceLen int
}

func (r *singleSearchReplaceReplacer) GetSizingHints() (int, int, float64) {
	r.searchLen = len(r.search)
	r.replaceLen = len(r.replace)
	ratio := float64(-1)
	if r.searchLen < r.replaceLen {
		ratio = float64(r.searchLen) / float64(r.replaceLen)
	}
	for i := 0; i < 24; i++ {
		r.slide[i]--
	}
	for i := 0; i < len(r.search); i++ {
		r.slide[(r.search)[i]] = i
	}
	return r.searchLen, r.replaceLen, ratio
}


// BestIndex Finds the best indexing option for `r.search` in `buf`, and runs it.
// If `len(r.search) == 1`, then `bytes.IndexByte()` is used. (SIMD ASM implementation)
// If `len(r.search) == 0`, then we can assume `r.search` is nowhere and everywhere. (returns 0)
// If `len(r.search) > len(buf)`, then we can assume that `r.search` does not exist under any circumstances. (returns -1)
// If `len(buf) == len(r.search)`, then we can assume the plausibility of `buf == r.search`. From here, we can check it using `bytes.Equal()`. (returns 0 if true)
// If all other conditions are untrue, then the Boyer-Moore indexing algorithm is used to get the position of `r.search` inside `buf`.
// Returns -1 if `r.search` is not contained within `buf`.
func (r *singleSearchReplaceReplacer) BestIndex(buf []byte) (int, []byte, []byte) {
	switch {
	case r.searchLen == 1:
		return bytes.IndexByte(buf, r.search[0]), r.search, r.replace
	case r.searchLen == 0:
		return 0, r.search, r.replace
	case r.searchLen > len(buf):
		fallthrough
	case len(buf) == 0:
		return -1, r.search, r.replace
	case len(buf) == r.searchLen:
		if bytes.Equal(buf, r.search) {
			return 0, r.search, r.replace
		}
		return -1, r.search, r.replace
	default:
		for i := 0; i+r.searchLen-1 < len(buf); {
			j := r.searchLen - 1
			for ; j >= 0 && buf[i+j] == r.search[j]; j-- {}
			if j < 0 {
				return i, r.search, r.replace
			}
			slid := j - r.slide[buf[i+j]]
			if slid < 1 {
				slid = 1
			}
			i += slid
		}
		return -1, r.search, r.replace
		//return r.BoyerMooreIndex(buf, r.search), r.search, r.replace
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// NewBytesReplacingReader creates a new `*BytesReplacingReader` for a single pair of search:replace token replacement.
// `search` cannot be nil/empty. `replace` can.
func NewBytesReplacingReader(r io.Reader, search, replace []byte) *BytesReplacingReader {
	return (&BytesReplacingReader{}).ResetEx(r, &singleSearchReplaceReplacer{search: search, replace: replace})
}

// NewBytesReplacingReaderEx creates a new `*BytesReplacingReader` for a given BytesReplacer customization.
func NewBytesReplacingReaderEx(r io.Reader, replacer BytesReplacer) *BytesReplacingReader {
	return (&BytesReplacingReader{}).ResetEx(r, replacer)
}

