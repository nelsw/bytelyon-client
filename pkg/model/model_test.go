package model

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"testing"
)

// encode gzips and base64 encodes s, as the scripts do.
func encode(t *testing.T, s string) string {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write([]byte(s)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func TestPageBytes(t *testing.T) {
	p := Page{Screenshot: encode(t, "png"), Content: encode(t, "<html>")}
	if got := string(p.ScreenshotBytes()); got != "png" {
		t.Errorf("ScreenshotBytes() = %q", got)
	}
	if got := string(p.ContentBytes()); got != "<html>" {
		t.Errorf("ContentBytes() = %q", got)
	}

	full := encode(t, "truncated")
	raw, _ := base64.StdEncoding.DecodeString(full)
	for name, s := range map[string]string{
		"invalid base64": "%%%",
		"not gzip":       base64.StdEncoding.EncodeToString([]byte("plain")),
		"truncated gzip": base64.StdEncoding.EncodeToString(raw[:len(raw)-4]),
	} {
		if got := (&Page{Content: s}).ContentBytes(); got != nil {
			t.Errorf("%s: ContentBytes() = %q, want nil", name, got)
		}
	}
}
