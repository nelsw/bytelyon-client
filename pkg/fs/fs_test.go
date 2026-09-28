package fs_test

import (
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
