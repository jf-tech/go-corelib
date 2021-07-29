package ios

import (
	"bufio"
	"io"

	"github.com/carterpeel/go-corelib/strs"
)

// ScannerByDelimFlag is the type of flags passed to NewScannerByDelim/NewScannerByDelim2.
type ScannerByDelimFlag uint

const (
	// ScannerByDelimFlagEofAsDelim specifies that the scanner should treat EOF as the delimiter as well.
	ScannerByDelimFlagEofAsDelim ScannerByDelimFlag = 1 << iota
	// ScannerByDelimFlagDropDelimInReturn specifies that the delimiter should NOT be included in the return value.
	ScannerByDelimFlagDropDelimInReturn
	scannerByDelimFlagEnd

	// ScannerByDelimFlagEofNotAsDelim specifies that the scanner should NOT treat EOF as the delimiter.
	ScannerByDelimFlagEofNotAsDelim = 0
	// ScannerByDelimFlagIncludeDelimInReturn specifies that the delimiter should be included in the return value.
	ScannerByDelimFlagIncludeDelimInReturn = 0
)
const (
	// ScannerByDelimFlagDefault specifies the most commonly used flags for the scanner.
	ScannerByDelimFlagDefault = ScannerByDelimFlagEofAsDelim | ScannerByDelimFlagDropDelimInReturn
	scannerByDelimValidFlags  = scannerByDelimFlagEnd - 1
)

// NewScannerByDelim creates a scanner that returns tokens from the source reader separated by a delimiter.
func NewScannerByDelim(r io.Reader, delim []byte, flags ScannerByDelimFlag) *bufio.Scanner {
	return NewScannerByDelim2(r, delim, nil, flags)
}

// NewScannerByDelim2 creates a scanner that returns tokens from the source reader separated by a delimiter, with
// consideration of potential presence of escaping sequence.
// Note: the token returned from the scanner will **NOT** do any unescaping, thus keeping the original value.
func NewScannerByDelim2(r io.Reader, delim, escape []byte, flags ScannerByDelimFlag) *bufio.Scanner {
	if escape == nil {
		return NewScannerByDelim3(r, delim, nil, flags, nil)
	}
	return NewScannerByDelim3(r, delim, escape, flags, nil)
}

// NewScannerByDelim3 creates a scanner that utilizes given buf to avoid/minimize allocation and returns tokens
// from the source reader separated by a delimiter, with consideration of potential presence of escaping sequence.
// Note: the token returned from the scanner will **NOT** do any unescaping, thus keeping the original value.
func NewScannerByDelim3(r io.Reader, delim, escape []byte, flags ScannerByDelimFlag, buf []byte) *bufio.Scanner {
	flags &= scannerByDelimValidFlags

	includeDelimLenInToken := len(delim)
	if flags&ScannerByDelimFlagDropDelimInReturn != 0 {
		includeDelimLenInToken = 0
	}
	eofAsDelim := flags&ScannerByDelimFlagEofAsDelim != 0

	scanner := bufio.NewScanner(r)
	scanner.Buffer(buf, bufio.MaxScanTokenSize)
	scanner.Split(
		func(data []byte, atEof bool) (advance int, token []byte, err error) {
			if atEof && len(data) == 0 {
				return 0, nil, nil
			}
			if index := strs.ByteIndexWithEsc(data, delim, escape); index >= 0 {
				return index + len(delim), data[:index+includeDelimLenInToken], nil
			}
			if atEof && eofAsDelim {
				return len(data), data, nil
			}
			return 0, nil, nil
		})
	return scanner
}
