package http

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/nelsw/bytelyon-client/pkg/url"
)

// echo responds with the request method, content type, authorization, and body so callers can assert on them.
func echo(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	_, _ = io.WriteString(w, r.Method+"|"+r.Header.Get("Content-Type")+"|"+r.Header.Get("Authorization")+"|"+string(b))
}

func TestRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(echo))
	defer srv.Close()

	tests := []struct {
		name string
		res  Response
		want string
	}{
		{"get", New(srv.URL).Bearer("tok").Get(), "GET||Bearer tok|"},
		{"put string", New(srv.URL).Put("raw"), "PUT|application/json||raw"},
		{"post bytes", New(srv.URL).Header("Content-Type", "text/plain").Post([]byte("b")), "POST|text/plain||b"},
		{"post json", New(srv.URL).Post(map[string]int{"a": 1}), `POST|application/json||{"a":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.res.Err != nil {
				t.Fatal(tt.res.Err)
			}
			if got := string(tt.res.Body); got != tt.want {
				t.Errorf("body = %q, want %q", got, tt.want)
			}
			if tt.res.URL != srv.URL {
				t.Errorf("url = %q, want %q", tt.res.URL, srv.URL)
			}
		})
	}
}

func TestRequestErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	if res := New(srv.URL).Get(); res.Err == nil || res.Err.Error() != "HTTP error 404" {
		t.Errorf("status error = %v", res.Err)
	}
	if res := New(srv.URL).Post(make(chan int)); res.Err == nil {
		t.Error("expected marshal error")
	}
	if res := New("://bad").Get(); res.Err == nil {
		t.Error("expected request error")
	}
	if res := New("http://127.0.0.1:1").Get(); res.Err == nil {
		t.Error("expected connection error")
	}
}

type failReader struct{}

func (failReader) Read([]byte) (int, error) { return 0, errors.New("boom") }
func (failReader) Close() error             { return nil }

func TestMakeResponseReadError(t *testing.T) {
	if res := makeResponse(failReader{}, http.StatusOK, "u"); res.Err == nil {
		t.Error("expected read error")
	}
}

func TestResponseXML(t *testing.T) {
	type doc struct {
		XMLName xml.Name `xml:"doc"`
		Value   string   `xml:"value"`
	}

	d, err := Response{Body: []byte("<doc><value>v</value></doc>")}.XML[doc]()
	if err != nil || d.Value != "v" {
		t.Errorf("XML() = %+v, %v", d, err)
	}

	if _, err = (Response{Err: errors.New("prior")}).XML[doc](); err == nil || err.Error() != "prior" {
		t.Errorf("XML() should propagate the response error, got %v", err)
	}
	if _, err = (Response{Body: []byte("<doc>")}).XML[doc](); err == nil {
		t.Error("XML() should fail on malformed xml")
	}
}

func TestGetRetriesTooManyRequests(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	b, err := Get(srv.URL+"/p?q=1", http.Header{"X-Test": {"1"}})
	if err != nil || string(b) != "ok" {
		t.Errorf("Get() = %q, %v", b, err)
	}
	if calls.Load() != 3 {
		t.Errorf("calls = %d, want 3", calls.Load())
	}
}

func TestGetGivesUp(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "nope", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	if _, err := Get(srv.URL, nil); err == nil || !strings.Contains(err.Error(), "[429] nope") {
		t.Errorf("Get() error = %v", err)
	}
	if calls.Load() != retryAttemptMax {
		t.Errorf("calls = %d, want %d", calls.Load(), retryAttemptMax)
	}
}

func TestGetNoRetryOnOtherErrors(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "gone", http.StatusGone)
	}))
	defer srv.Close()

	if _, err := Get(srv.URL, nil); err == nil {
		t.Error("expected error")
	}
	if calls.Load() != 1 {
		t.Errorf("calls = %d, want 1", calls.Load())
	}
}

func TestGetNode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "<html><body><p>hi</p></body></html>")
	}))
	defer srv.Close()

	n, err := GetNode(srv.URL)
	if err != nil || n == nil || n.FirstChild == nil {
		t.Fatalf("GetNode() = %v, %v", n, err)
	}

	if _, err = GetNode("http://127.0.0.1:1"); err == nil {
		t.Error("expected connection error")
	}
}

func TestDo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(echo))
	defer srv.Close()

	tests := []struct {
		name   string
		method string
		body   any
		header http.Header
		want   string
	}{
		{"reader body", http.MethodPut, strings.NewReader("r"), nil, "PUT|application/json||r"},
		{"json body", http.MethodPost, []int{1}, http.Header{"Content-Type": {"x/y"}}, "POST|x/y||[1]"},
		{"no body", http.MethodDelete, nil, nil, "DELETE|||"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, code, err := do(tt.method, srv.URL, tt.body, tt.header)
			if err != nil || code != http.StatusOK || string(b) != tt.want {
				t.Errorf("do() = %q, %d, %v; want %q", b, code, err, tt.want)
			}
		})
	}

	if _, _, err := do(http.MethodPost, srv.URL, make(chan int), nil); err == nil {
		t.Error("expected marshal error")
	}
	if _, _, err := do(http.MethodGet, "://bad", nil, nil); err == nil {
		t.Error("expected request error")
	}
}

func TestDoReadError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10")
		_, _ = io.WriteString(w, "short")
	}))
	defer srv.Close()

	if _, _, err := do(http.MethodGet, srv.URL, nil, nil); err == nil {
		t.Error("expected read error for truncated body")
	}
}

func TestPostForm(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, _ = io.WriteString(w, r.Header.Get("Content-Type")+"|"+r.UserAgent()+"|"+r.PostForm.Get("f"))
	}))
	defer srv.Close()

	b, err := PostForm(srv.URL, url.Values{"f": {"v"}})
	if want := "application/x-www-form-urlencoded;charset=UTF-8|" + userAgent + "|v"; err != nil || string(b) != want {
		t.Errorf("PostForm() = %q, %v", b, err)
	}

	if _, err = PostForm("://bad", nil); err == nil {
		t.Error("expected request error")
	}
	if _, err = PostForm("http://127.0.0.1:1", nil); err == nil {
		t.Error("expected connection error")
	}
}
