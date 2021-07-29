package ios

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/carterpeel/go-corelib/testlib"
)

func TestFileExists(t *testing.T) {
	// non existing case.
	assert.False(t, FileExists("non-existing"))
	tmp := testlib.CreateTempFileWithContent(t, "", "", "test")
	defer os.Remove(tmp.Name())
	// file existing case.
	assert.True(t, FileExists(tmp.Name()))
	// existing but not a file case.
	assert.False(t, FileExists(filepath.Dir(tmp.Name())))
}

func TestDirExists(t *testing.T) {
	// non existing case.
	assert.False(t, DirExists("non-existing"))
	tmp := testlib.CreateTempFileWithContent(t, "", "", "test")
	defer os.Remove(tmp.Name())
	// dir existing case.
	assert.True(t, DirExists(filepath.Dir(tmp.Name())))
	// existing but not a dir case.
	assert.False(t, DirExists(tmp.Name()))
}
