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
)

// BuildFQDN builds an FQDN from a slice of namelet strings.
func BuildFQDN(namelets ...string) string {
	return BuildFQDN2(FQDNDelimiter, namelets...)
}

// BuildFQDN2 builds an FQDN from a slice of namelet strings and a given delimiter.
func BuildFQDN2(delimiter string, namelets ...string) string {
	return strings.Join(namelets, delimiter)
}

// LastNameletOfFQDN returns the last namelet of an FQDN delimited by default
// delimiter. If there is no delimiter in the FQDN, then the FQDN itself is
// // returned.
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

// IndexWithEsc is similar to strings.Index but taking escape sequence into consideration.
// For example, IndexWithEsc("abc%|efg|xyz", "|", RunePtr("%")) would return 8, not 4.
func IndexWithEsc(s, delim string, esc *rune) int {
	if len(delim) == 0 {
		return 0
	}
	if len(s) == 0 {
		return -1
	}
	if esc == nil {
		return strings.Index(s, delim)
	}
	sRunes := []rune(s)
	delimRunes := []rune(delim)
	escRune := *esc

	// Yes this old dumb double loop isn't the most efficient algo but it's super easy and simple to understand
	// and bug free compared with fancy strings.Index or bytes.Index which could potentially lead to index errors
	// and/or rune/utf-8 bugs. Plus for vast majority of use cases, delim will be of a single rune, so effectively
	// not much perf penalty at all.
	for i := 0; i < len(sRunes)-len(delimRunes)+1; i++ {
		if sRunes[i] == escRune {
			// skip the escaped rune (aka the rune after the escape rune)
			i++
			continue
		}
		sIndex := i
		delimIndex := 0
		for sIndex < len(sRunes) && delimIndex < len(delimRunes) {
			if sRunes[sIndex] == escRune {
				sIndex += 2
				continue
			}
			if sRunes[sIndex] != delimRunes[delimIndex] {
				break
			}
			sIndex++
			delimIndex++
		}
		if delimIndex >= len(delimRunes) {
			return len(string(sRunes[:i]))
		}
	}

	return -1
}

// SplitWithEsc is similar to strings.Split but taking escape sequence into consideration.
// For example, SplitWithEsc("abc%|efg|xyz", "|", RunePtr("%")) would return []string{"abc%|efg", "xyz"}.
func SplitWithEsc(s, delim string, esc *rune) []string {
	if len(delim) == 0 || esc == nil {
		return strings.Split(s, delim)
	}
	// From here on, delim != empty **and** esc is set.
	var split []string
	for delimIndex := IndexWithEsc(s, delim, esc); delimIndex >= 0; delimIndex = IndexWithEsc(s, delim, esc) {
		split = append(split, s[:delimIndex])
		s = s[delimIndex+len(delim):]
	}
	split = append(split, s)
	return split
}

// ByteSplitWithEsc is similar to SplitWithEsc but operating on []byte and []rune. Also esc is not optional.
// cap is just an indicator to the function that what the initial cap the returned slice should be pre-allocated.
func ByteSplitWithEsc(s []byte, delim []rune, esc rune, cap int) [][]byte {
	if len(delim) == 0 {
		return bytes.Split(s, nil)
	}
	delimByteLen := 0
	for i := 0; i < len(delim); i++ {
		delimByteLen += utf8.RuneLen(delim[i])
	}
	type runeWithIndex struct {
		r         rune
		byteIndex int
	}
	// to avoid repeatedly calling utf8.DecodeRune, let's just take the hit and do a memory allocation
	// and convert s into runes once.
	src := make([]runeWithIndex, 0, len(s))
	for i := 0; ; {
		r, size := utf8.DecodeRune(s[i:])
		if r == utf8.RuneError {
			break
		}
		src = append(src, runeWithIndex{r: r, byteIndex: i})
		i += size
	}
	findDelim := func(src []runeWithIndex) int {
		for i := 0; i < len(src)-len(delim)+1; i++ {
			if src[i].r == esc {
				i++
				continue
			}
			srcIndex := i
			delimIndex := 0
			for srcIndex < len(src) && delimIndex < len(delim) {
				if src[srcIndex].r == esc {
					srcIndex += 2
					continue
				}
				if src[srcIndex].r != delim[delimIndex] {
					break
				}
				srcIndex++
				delimIndex++
			}
			if delimIndex >= len(delim) {
				return i
			}
		}
		return -1
	}
	splits := make([][]byte, 0, cap)
	splitBegin := 0
	for delimIndexInSrc := findDelim(src); delimIndexInSrc >= 0; delimIndexInSrc = findDelim(src) {
		splits = append(splits, s[splitBegin:src[delimIndexInSrc].byteIndex])
		splitBegin = src[delimIndexInSrc].byteIndex + delimByteLen
		src = src[delimIndexInSrc+len(delim):]
	}
	splits = append(splits, s[splitBegin:])
	return splits
}

// Unescape unescapes a string with escape sequence.
// For example, SplitWithEsc("abc%|efg", RunePtr("%")) would return "abc|efg".
func Unescape(s string, esc *rune) string {
	if esc == nil {
		return s
	}
	sRunes := []rune(s)
	escRune := *esc
	for i := 0; i < len(sRunes); i++ {
		if sRunes[i] != escRune {
			continue
		}
		copy(sRunes[i:], sRunes[i+1:])
		sRunes = sRunes[:len(sRunes)-1]
	}
	return string(sRunes)
}
