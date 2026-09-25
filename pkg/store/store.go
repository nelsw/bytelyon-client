package store

import (
	"fmt"
	"os"
	"path/filepath"
)

const storagePath = ".storage"

func init() {
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		panic(err)
	}
}

func Mkdirs(aa ...any) string {

	if len(aa) == 0 {
		return storagePath
	}

	parts := []string{storagePath}
	for _, a := range aa {
		parts = append(parts, fmt.Sprint(a))
	}

	_ = os.MkdirAll(filepath.Join(parts[:len(parts)-1]...), 0755)
	return filepath.Join(parts...)
}

func Create(aa ...any) (f *os.File, err error) {
	return os.Create(Mkdirs(aa...))
}

func Read(aa ...any) ([]byte, error) {
	path := storagePath
	for _, p := range aa {
		path = filepath.Join(path, fmt.Sprint(p))
	}
	return os.ReadFile(path)
}
