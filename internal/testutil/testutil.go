// Package testutil wires the app's external dependencies (Postgres, Redis, S3, SSH, Playwright scripts)
// to in-process fakes so tests never touch real infrastructure, even when run with a populated .env.
package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/nelsw/bytelyon-client/pkg/cache"
)

// S3 is a fake S3 endpoint that records the objects put to it.
type S3 struct {
	*httptest.Server
	mu      sync.Mutex
	objects map[string][]byte
	fail    bool
}

// Object returns the body stored at key (path-style, e.g. "bucket/key") and whether it exists.
func (s *S3) Object(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.objects[key]
	return b, ok
}

// Fail makes subsequent requests respond with a (non-retryable) access denied error.
func (s *S3) Fail(fail bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fail = fail
}

func (s *S3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail {
		http.Error(w, "<Error><Code>AccessDenied</Code></Error>", http.StatusForbidden)
		return
	}
	s.objects[strings.TrimPrefix(r.URL.Path, "/")] = b
	w.Header().Set("ETag", `"etag"`)
}

var (
	s3Once sync.Once
	s3Fake *S3
)

// Bucket is the S3 bucket name tests run against.
const Bucket = "test-bucket"

// Isolate points every external dependency at a safe, local default. It uses os.Setenv rather than t.Setenv
// because the fs and cache packages read their configuration lazily, possibly after the test that set it ends.
func Isolate() *S3 {
	s3Once.Do(func() {
		s3Fake = &S3{objects: map[string][]byte{}}
		s3Fake.Server = httptest.NewServer(s3Fake)
		for k, v := range map[string]string{
			"SERVER_ADDR":                 "",
			"POSTGRES_URL":                "postgres://test@127.0.0.1:1/test?connect_timeout=1",
			"REDIS_ADDR":                  "127.0.0.1:1",
			"REDIS_DB":                    "0",
			"S3_BUCKET":                   Bucket,
			"AWS_ENDPOINT_URL_S3":         s3Fake.URL,
			"AWS_REGION":                  "us-east-1",
			"AWS_ACCESS_KEY_ID":           "test",
			"AWS_SECRET_ACCESS_KEY":       "test",
			"AWS_CONFIG_FILE":             os.DevNull,
			"AWS_SHARED_CREDENTIALS_FILE": os.DevNull,
			"AWS_PROFILE":                 "",
			"AWS_EC2_METADATA_DISABLED":   "true",
		} {
			_ = os.Setenv(k, v)
		}
	})
	return s3Fake
}

// Redis starts a miniredis server and points the cache at it for the duration of the test.
func Redis(t testing.TB) *miniredis.Miniredis {
	t.Helper()
	Isolate()
	cache.Close()
	m := miniredis.RunT(t)
	_ = os.Setenv("REDIS_ADDR", m.Addr())
	t.Cleanup(func() {
		cache.Close()
		_ = os.Setenv("REDIS_ADDR", "127.0.0.1:1")
	})
	return m
}

// Workdir changes into a fresh temporary directory, which receives the .storage and scripts trees.
func Workdir(t testing.TB) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	return dir
}

// Script installs an executable ./scripts/<name> that records its arguments to <name>.args and exits with code.
func Script(t testing.TB, name string, code int) {
	t.Helper()
	body := "#!/bin/sh\necho \"$@\" > \"" + name + ".args\"\nexit " + strconv.Itoa(code) + "\n"
	if err := os.MkdirAll("scripts", 0o755); err != nil {
		t.Fatal(err)
	} else if err = os.WriteFile(filepath.Join("scripts", name), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

// Args returns the arguments the named script was last invoked with.
func Args(t testing.TB, name string) string {
	t.Helper()
	b, err := os.ReadFile(name + ".args")
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}

// Files writes the <path>.html, <path>.png, and <path>.json files a Playwright script would produce.
func Files(t testing.TB, path, json string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	for ext, body := range map[string]string{".html": "<html></html>", ".png": "\x89PNG\r\n\x1a\n", ".json": json} {
		if err := os.WriteFile(path+ext, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// Transport routes every request made through http.DefaultTransport to handler for the duration of the test.
func Transport(t testing.TB, handler http.Handler) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	orig := http.DefaultTransport
	http.DefaultTransport = &rewrite{target: srv.Listener.Addr().String(), base: orig}
	t.Cleanup(func() { http.DefaultTransport = orig })
}

type rewrite struct {
	target string
	base   http.RoundTripper
}

func (r *rewrite) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("X-Original-Host", req.URL.Host)
	req.URL.Scheme = "http"
	req.URL.Host = r.target
	return r.base.RoundTrip(req)
}
