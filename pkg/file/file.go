package file

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var absPath string

func init() { absPath, _ = os.Executable() }

func MakeDir(path string) error            { return os.MkdirAll(path, 0755) }
func Read(path string) ([]byte, error)     { return os.ReadFile(path) }
func Remove(path string) error             { return os.Remove(path) }
func Write(path string, data []byte) error { return os.WriteFile(path, data, 0644) }

func ReadType[T any](path string, absolutePath ...bool) (t T, err error) {
	if len(absolutePath) > 0 && absolutePath[0] {
		path = filepath.Join(filepath.Dir(absPath), path)
	}
	var b []byte
	if b, err = Read(path); err == nil {
		err = json.Unmarshal(b, &t)
	}
	return
}

func ReadStr(path string, absolutePath ...bool) (string, error) {
	if len(absolutePath) > 0 && absolutePath[0] {
		path = filepath.Join(filepath.Dir(absPath), path)
	}
	b, err := Read(path)
	return string(b), err
}
