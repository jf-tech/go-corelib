package strs

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

// RunePtr returns a pointer to a rune.
func RunePtr(r rune) *rune {
	return &r
}

// StrPtr returns string pointer that points to a given string value.
func StrPtr(s string) *string {
	return &s
}

// IsStrNonBlank checks if a string is blank or not.
func IsStrNonBlank(s string) bool {
	return len(strings.TrimSpace(s)) > 0
}

// IsStrPtrNonBlank checks if the value represented by a string pointer is blank or not.
func IsStrPtrNonBlank(sp *string) bool { return sp != nil && IsStrNonBlank(*sp) }

// FirstNonBlank returns the first non-blank string value of the input strings, if any; or "" is returned.
func FirstNonBlank(strs ...string) string {
	for _, str := range strs {
		if IsStrNonBlank(str) {
			return str
		}
	}
	return ""
}

// StrPtrOrElse returns the string value of the string pointer if non-nil, or the default string value.
func StrPtrOrElse(sp *string, orElse string) string {
	if sp != nil {
		return *sp
	}
	return orElse
}

// CopyStrPtr copies a string pointer and its underlying string value, if set, into a new string pointer.
func CopyStrPtr(sp *string) *string {
	if sp == nil {
		return nil
	}
	s := *sp
	return &s
}

const (
	// FQDNDelimiter is the default FQDN delimiter.
	FQDNDelimiter = "."
	// FQDNEsc is the default escape char for FQDN. Esc is used for escaping "." and itself.
	FQDNEsc = "%"
)

// BuildFQDN builds an FQDN (dot delimited) from a slice of namelet strings.
func BuildFQDN(namelets ...string) string {
	return BuildFQDN2(FQDNDelimiter, namelets...)
}

// BuildFQDN2 builds an FQDN from a slice of namelet strings and a given delimiter.
func BuildFQDN2(delimiter string, namelets ...string) string {
	return strings.Join(namelets, delimiter)
}

var fqdnEscReplacer = strings.NewReplacer(".", "%.", "%", "%%")

// BuildFQDNWithEsc builds an FQDN (dot delimited) from a slice of namelet strings with proper escaping.
// e.g. If namelets are 'a.b', 'c%d', it will return 'a%.b.c%%d'.
// Note this function isn't optimized for alloc/perf.
func BuildFQDNWithEsc(namelets ...string) string {
	return strings.Join(
		NoErrMapSlice(namelets, func(s string) string {
			return fqdnEscReplacer.Replace(s)
		}),
		FQDNDelimiter)
}

// LastNameletOfFQDN returns the last namelet of an FQDN delimited by default
// delimiter. If there is no delimiter in the FQDN, then the FQDN itself is
// returned.
func LastNameletOfFQDN(fqdn string) string {
	return LastNameletOfFQDN2(FQDNDelimiter, fqdn)
}

// LastNameletOfFQDN2 returns the last namelet of an FQDN delimited by given
// delimiter. If there is no delimiter in the FQDN, then the FQDN itself is
// returned.
func LastNameletOfFQDN2(delimiter, fqdn string) string {
	index := strings.LastIndex(fqdn, delimiter)
	if index < 0 {
		return fqdn
	}
	return fqdn[index+1:]
}

// LastNameletOfFQDNWithEsc returns the last namelet of an FQDN delimited by default
// delimiter, with escaping considered. If there is no delimiter in the FQDN, then the
// FQDN itself is returned.
// Note this function isn't optimized for alloc/perf.
func LastNameletOfFQDNWithEsc(fqdn string) string {
	namelets := SplitWithEsc(fqdn, FQDNDelimiter, FQDNEsc)
	return Unescape(namelets[len(namelets)-1], FQDNEsc)
}

// CopySlice copies a string slice. The returned slice is guaranteed to be a different
// slice (thus the name Copy) so modifying the src from the caller side won't affect
// the returned slice.
func CopySlice(src []string) []string {
	return MergeSlices(src, nil)
}

// MergeSlices returns a new slice with two input slice content merged together. The result
// is guaranteed to be a new slice thus modifying a or b from the caller side won't affect
// the returned slice.
func MergeSlices(a, b []string) []string {
	return append(append([]string(nil), a...), b...)
}

// HasDup detects whether there are duplicates existing in the src slice.
func HasDup(src []string) bool {
	seen := map[string]bool{}
	for _, v := range src {
		if _, found := seen[v]; found {
			return true
		}
		seen[v] = true
	}
	return false
}

// MapSlice returns a new string slice whose element is transformed from input slice's
// corresponding element by a transform func. If any error occurs during any transform,
// returned slice will be nil together with the error.
func MapSlice(src []string, f func(string) (string, error)) ([]string, error) {
	if len(src) == 0 {
		return nil, nil
	}
	result := make([]string, len(src))
	for i := 0; i < len(src); i++ {
		s, err := f(src[i])
		if err != nil {
			return nil, err
		}
		result[i] = s
	}
	return result, nil
}

// NoErrMapSlice returns a new string slice whose element is transformed from input slice's
// corresponding element by a transform func. The transform func must not fail and NoErrMapSlice
// guarantees to succeed.
func NoErrMapSlice(src []string, f func(string) string) []string {
	result, _ := MapSlice(src, func(s string) (string, error) {
		return f(s), nil
	})
	return result
}

// ByteLenOfRunes returns byte length of a rune slice.
func ByteLenOfRunes(rs []rune) int {
	byteLen := 0
	for i := 0; i < len(rs); i++ {
		byteLen += utf8.RuneLen(rs[i])
	}
	return byteLen
}

// IndexWithEsc searches for 'delim' inside 's' with escaping sequence taking into account.
// Note 'delim' must not contain 'esc', or if it does, 'esc' inside 'delim' will be treated as
// regular string.
func IndexWithEsc(s, delim string, esc string) int {
	if len(s) == 0 || len(delim) == 0 || len(esc) == 0 {
		return strings.Index(s, delim)
	}
	isEscPreceding := func(i int) bool {
		// this func check if there is an effective 'esc' directly preceding
		// the byte at i. it is not trivial to check the byte sequence right
		// before i, because we can have multiple 'esc' escaping each other.
		// so we need to backtrack (hopefully not for too far)
		escFound := 0
		for i >= len(esc) {
			if strings.Index(s[i-len(esc):i], esc) < 0 {
				break
			}
			escFound++
			i -= len(esc)
		}
		return escFound%2 == 1
	}
	begin := 0
	for {
		i := strings.Index(s[begin:], delim)
		if i < 0 {
			return i
		}
		begin += i
		// we've found the 'delim', looks like this:
		//   s [..............delim...........]
		//                    ^begin
		// However, we need to check if there is an effective 'esc' directly preceding 'delim'
		// or not. If yes, we will have to skip the first rune inside the 'delim' (because it
		// is escaped by the preceding 'esc') and redo the whole process.
		if isEscPreceding(begin) {
			// no need to check utf8.RuneError because we've come here because we've found 'delim'
			// at 'begin' and 'len(delim)' isn't 0, so there is at least one rune there at 'begin'.
			_, size := utf8.DecodeRuneInString(s[begin:])
			begin += size
			continue
		}
		return begin
	}
}

// SplitWithEsc is similar to strings.Split but taking escape sequence into consideration.
// For example, SplitWithEsc("abc%|efg|xyz", "|", "%") would return []string{"abc%|efg", "xyz"}.
func SplitWithEsc(s, delim string, esc string) []string {
	if len(s) == 0 || len(delim) == 0 || len(esc) == 0 {
		return strings.Split(s, delim)
	}
	var split []string
	for index := IndexWithEsc(s, delim, esc); index >= 0; index = IndexWithEsc(s, delim, esc) {
		split = append(split, s[:index])
		s = s[index+len(delim):]
	}
	split = append(split, s)
	return split
}

// ByteIndexWithEsc searches for 'delim' inside 's' with escaping sequence taking into account.
// Note 'delim' must not contain 'esc', or if it does, 'esc' inside 'delim' will be treated as
// regular bytes.
func ByteIndexWithEsc(s, delim, esc []byte) int {
	if len(s) == 0 || len(delim) == 0 || len(esc) == 0 {
		return bytes.Index(s, delim)
	}
	isEscPreceding := func(i int) bool {
		// this func check if there is an effective 'esc' directly preceding
		// the byte at i. it is not trivial to check the byte sequence right
		// before i, because we can have multiple 'esc' escaping each other.
		// so we need to backtrack (hopefully not for too far)
		escFound := 0
		for i >= len(esc) {
			if bytes.Index(s[i-len(esc):i], esc) < 0 {
				break
			}
			escFound++
			i -= len(esc)
		}
		return escFound%2 == 1
	}
	begin := 0
	for {
		i := bytes.Index(s[begin:], delim)
		if i < 0 {
			return i
		}
		begin += i
		// we've found the 'delim', looks like this:
		//   s [..............delim...........]
		//                    ^begin
		// However, we need to check if there is an effective 'esc' directly preceding 'delim'
		// or not. If yes, we will have to skip the first rune inside the 'delim' (because it
		// is escaped by the preceding 'esc') and redo the whole process.
		if isEscPreceding(begin) {
			// no need to check utf8.RuneError because we've come here because we've found 'delim'
			// at 'begin' and 'len(delim)' isn't 0, so there is at least one rune there at 'begin'.
			_, size := utf8.DecodeRune(s[begin:])
			begin += size
			continue
		}
		return begin
	}
}

// ByteSplitWithEsc is similar to SplitWithEsc but operating on []byte. 'cap' is just an indicator to
// the function that what the initial cap the returned split slice should be pre-allocated.
func ByteSplitWithEsc(s, delim, esc []byte, cap int) [][]byte {
	if len(s) == 0 || len(delim) == 0 || len(esc) == 0 {
		return bytes.Split(s, delim)
	}
	splits := make([][]byte, 0, cap)
	for index := ByteIndexWithEsc(s, delim, esc); index >= 0; index = ByteIndexWithEsc(s, delim, esc) {
		splits = append(splits, s[:index])
		s = s[index+len(delim):]
	}
	splits = append(splits, s)
	return splits
}

// Unescape unescapes a string with escape sequence.
// For example, SplitWithEsc("abc%|efg", "%") would return "abc|efg".
func Unescape(s, esc string) string {
	if len(esc) == 0 {
		return s
	}
	sb := strings.Builder{}
	sb.Grow(len(s))
	for {
		i := strings.Index(s, esc)
		if i < 0 {
			sb.WriteString(s)
			break
		}
		sb.WriteString(s[:i])
		r, size := utf8.DecodeRuneInString(s[i+len(esc):])
		if r == utf8.RuneError {
			break
		}
		sb.WriteRune(r)
		s = s[i+len(esc)+size:]
	}
	return sb.String()
}

// ByteUnescape unescapes a []byte sequence with an escape []byte sequence.
// If 'inPlace' is true, the unescaping modification is taking place inside
// the original 'b' and 'b' will be returned with length properly adjusted.
// If 'inPlace' is false, the unescaping is taking place inside a new []byte,
// the new []byte will be returned and the original 'b' is untouched.
func ByteUnescape(b, esc []byte, inPlace bool) []byte {
	if b == nil {
		return nil
	}
	result := b
	if !inPlace {
		result = make([]byte, len(b))
	}
	count := 0
	copyToResult := func(src []byte) {
		// because unescaping is strictly shrinking the slice thus
		// this copy algo is safe when inPlace is true as well.
		for i := 0; i < len(src); i++ {
			result[count+i] = src[i]
		}
		count += len(src)
	}
	if len(esc) == 0 {
		if inPlace {
			return b
		}
		copyToResult(b)
		return result[:count]
	}
	for {
		i := bytes.Index(b, esc)
		if i < 0 {
			copyToResult(b)
			break
		}
		copyToResult(b[:i])
		r, size := utf8.DecodeRune(b[i+len(esc):])
		if r == utf8.RuneError {
			break
		}
		copyToResult(b[i+len(esc) : i+len(esc)+size])
		b = b[i+len(esc)+size:]
	}
	return result[:count]
}
