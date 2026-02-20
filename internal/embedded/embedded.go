package embedded

import (
	"fmt"
	"io/fs"
)

func ReadFile(path string) ([]byte, error) {
	return nil, fmt.Errorf("embedded data not available: %s", path)
}

func ReadDir(path string) ([]fs.DirEntry, error) {
	return nil, fmt.Errorf("embedded data not available: %s", path)
}

func ReadDirFS(path string) ([]fs.DirEntry, error) {
	return nil, fmt.Errorf("embedded data not available: %s", path)
}

func WalkDir(root string, fn fs.WalkDirFunc) error {
	return fmt.Errorf("embedded data not available: %s", root)
}
