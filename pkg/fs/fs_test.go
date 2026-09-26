package fs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nelsw/bytelyon-client/internal/testutil"
	"github.com/nelsw/bytelyon-client/pkg/fs"
)

func TestPut(t *testing.T) {
	s3 := testutil.Isolate()

	if err := fs.Put("a/b.txt", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	if b, ok := s3.Object(testutil.Bucket + "/a/b.txt"); !ok || string(b) != "hello" {
		t.Errorf("object = %q, %v", b, ok)
	}
}

func TestPutErrors(t *testing.T) {
	s3 := testutil.Isolate()

	if err := fs.Put("", []byte("x")); err == nil {
		t.Error("expected empty key error")
	}
	if err := fs.Put("k", nil); err == nil {
		t.Error("expected empty data error")
	}

	s3.Fail(true)
	defer s3.Fail(false)
	if err := fs.Put("k", []byte("x")); err == nil {
		t.Error("expected server error")
	}
}

func TestMove(t *testing.T) {
	s3 := testutil.Isolate()
	from := filepath.Join(t.TempDir(), "page.html")
	if err := os.WriteFile(from, []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := fs.Move(from, "moved/page.html"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(from); !os.IsNotExist(err) {
		t.Errorf("source should be removed, stat err = %v", err)
	}
	if b, ok := s3.Object(testutil.Bucket + "/moved/page.html"); !ok || string(b) != "<html></html>" {
		t.Errorf("object = %q, %v", b, ok)
	}
}

func TestMoveErrors(t *testing.T) {
	testutil.Isolate()
	dir := t.TempDir()

	if err := fs.Move(filepath.Join(dir, "missing"), "k"); err == nil {
		t.Error("expected read error")
	}

	empty := filepath.Join(dir, "empty")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := fs.Move(empty, "k"); err == nil {
		t.Error("expected put error for empty file")
	}
	if _, err := os.Stat(empty); err != nil {
		t.Errorf("source should be kept when the upload fails: %v", err)
	}
}
