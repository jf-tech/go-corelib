package ios

import (
	"bytes"
	"io"

	"github.com/jf-tech/go-corelib/maths"
)

// LineEditFunc edits a line and returns a resulting line. Note in-place editing is highly encouraged,
// for performance reasons, when the resulting line is no longer than the original. If your edited line
// is longer then the original `line`, however, you MUST allocate and return a new []byte. Directly
// appending at the end of the original `line` will result in undefined behavior.
type LineEditFunc func(line []byte) ([]byte, error)

// LineReader implements io.Reader interface with a line editing mechanism. LineReader reads data from
// underlying io.Reader and invokes the caller supplied edit function for each of the line (defined as
// []byte ending with '\n', therefore it works on both Mac/Linux and Windows, where '\r\n' is used).
// Note the last line before EOF will be edited as well even if it doesn't end with '\n'. Usage is highly
// flexible: the editing function can do in-place editing such as character replacement, prefix/suffix
// stripping, or word replacement, etc., as long as the line length isn't changed; or it can replace a line
// with a completely newly allocated and written line with no length restriction (although performance
// would be slower compared to in-place editing).
type LineReader struct {
	r       io.Reader
	edit    LineEditFunc
	bufSize int    // initial buf size and future buf growth increment.
	buf     []byte // note len(buf) == cap(buf), we always use the full capacity of the buf.
	buf0    int    // buf[:buf0] edited line(s) ready to be returned to caller.
	buf1    int    // buf[buf0:buf1] unedited lines.
	err     error
}

func (r *LineReader) scanLF(buf []byte) int {
	if lf := bytes.IndexByte(buf, '\n'); lf >= 0 {
		return lf
	}
	if r.err == io.EOF {
		return len(buf) - 1
	}
	return -1
}

// Read implements io.Reader interface for LineReader.
func (r *LineReader) Read(p []byte) (int, error) {
	n := 0
	for {
		if r.buf0 > 0 {
			n = copy(p, r.buf[:r.buf0])
			r.buf0 -= n
			r.buf1 -= n
			copy(r.buf, r.buf[n:r.buf1+n])
			return n, nil
		} else if r.err != nil {
			return 0, r.err
		}

		if r.buf1 >= len(r.buf) {
			newBuf := make([]byte, len(r.buf)+r.bufSize)
			copy(newBuf, r.buf)
			r.buf = newBuf
		}

		n, r.err = r.r.Read(r.buf[r.buf1:])
		r.buf1 += n
		lf := r.scanLF(r.buf[r.buf0:r.buf1])
		for ; lf >= 0; lf = r.scanLF(r.buf[r.buf0:r.buf1]) {
			lineLen := lf + 1
			edited, err := r.edit(r.buf[r.buf0 : r.buf0+lineLen])
			if err != nil {
				r.err = err
				break
			}
			editedLen := len(edited)
			delta := lineLen - editedLen
			if len(r.buf)-r.buf1+delta < 0 {
				// only expand the buf if there is no room left for the edited line growth.
				newBuf := make([]byte, len(r.buf)+maths.MaxInt(r.bufSize, -delta))
				copy(newBuf, r.buf[:r.buf1])
				r.buf = newBuf
			}
			if delta > 0 {
				// This is the case where the edited line is shorter than the original line.
				// Image we have:
				//  xyz\nabc
				// where "xyz\n" is in-placed edited to drop the first letter to "yz\n".
				// If we shift "abc" up by delta (1) first, then we would've overwritten the "\n" in "yz\n"
				// and the edited would now be "yza".
				// Therefore, if edited is shorter, we need to move/copy edited to be at buf0 first
				// before we shift the rest of the buffer (up to buf1) up.
				copy(r.buf[r.buf0:r.buf0+editedLen], edited)
				copy(r.buf[r.buf0+editedLen:r.buf1-delta], r.buf[r.buf0+lineLen:r.buf1])
			} else {
				// Now if edited is longer, we need to move the rest buffer out first, before we can copy
				// the edited into the buffer.
				copy(r.buf[r.buf0+editedLen:r.buf1-delta], r.buf[r.buf0+lineLen:r.buf1])
				copy(r.buf[r.buf0:r.buf0+editedLen], edited)
			}
			r.buf0 += editedLen
			r.buf1 -= delta
		}
	}
}

// NewLineReader2 creates a new LineReader with custom buffer size.
func NewLineReader2(r io.Reader, edit LineEditFunc, bufSize int) *LineReader {
	buf := make([]byte, bufSize)
	return &LineReader{
		r:       r,
		edit:    edit,
		bufSize: bufSize,
		buf:     buf,
	}
}

const (
	defaultLineReaderBufSize = 1024
)

// NewLineReader creates a new LineReader with the default buffer size.
func NewLineReader(r io.Reader, edit LineEditFunc) *LineReader {
	return NewLineReader2(r, edit, defaultLineReaderBufSize)
}
