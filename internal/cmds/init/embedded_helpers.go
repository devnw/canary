package init

import (
	"io/fs"
)

// ReadFile exposes reading from the embedded filesystem
func ReadFile(path string) ([]byte, error) {
	return embedded.ReadFile(path)
}

// ReadDir exposes ReadDir on the embedded filesystem
func ReadDir(path string) ([]fs.DirEntry, error) {
	return embedded.ReadDir(path)
}

// WalkDir exposes fs.WalkDir over the embedded filesystem
func WalkDir(root string, fn fs.WalkDirFunc) error {
	return fs.WalkDir(embedded, root, fn)
}
