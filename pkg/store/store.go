package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nelsw/bytelyon-client/pkg/file"
)

const storagePath = ".storage"

func init() {
	if err := file.MakeDir(storagePath); err != nil {
		panic(err)
	}
}

func Make(aa []any) string {

	if len(aa) == 0 {
		return storagePath
	}

	parts := []string{storagePath}
	for _, a := range aa {
		parts = append(parts, fmt.Sprint(a))
	}

	_ = file.MakeDir(filepath.Join(parts[:len(parts)-1]...))
	return filepath.Join(parts...)
}

func Create(aa ...any) (f *os.File, err error) {
	return os.Create(Make(aa))
}

func Save(a any, aa ...any) (err error) {

	if a == nil {
		return fmt.Errorf("cannot save nil")
	}

	var b []byte
	switch a.(type) {
	case []byte:
		b = a.([]byte)
	case string:
		b = []byte(a.(string))
	default:
		if b, err = json.MarshalIndent(a, "", "\t"); err != nil {
			return
		}
	}

	return file.Write(Make(aa), b)
}

func Find[T any](aa ...any) (t T, err error) {
	path := storagePath
	for _, p := range aa {
		path = filepath.Join(path, fmt.Sprint(p))
	}
	var b []byte
	if b, err = file.Read(path); err == nil {
		err = json.Unmarshal(b, &t)
	}
	return
}
