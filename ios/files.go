package ios

import (
	"os"
)

// FileExists checks if a file exists or not.
func FileExists(file string) bool {
	fi, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	return !fi.IsDir()
}

// DirExists checks if a directory exists or not.
func DirExists(dir string) bool {
	fi, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}
	return fi.IsDir()
}
