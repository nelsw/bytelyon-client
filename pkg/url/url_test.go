package url

import (
	"maps"
	"testing"
)

func TestValuesEncode(t *testing.T) {
	v := Values{"b": {"2"}, "a": {"1 &"}}
	if got, want := string(v.Encode()), "a=1+%26&b=2"; got != want {
		t.Errorf("Encode() = %q, want %q", got, want)
	}
}

func TestClean(t *testing.T) {
	if got, want := Clean("  HTTPS://Example.com/Path/ "), "https://example.com/path"; got != want {
		t.Errorf("Clean() = %q, want %q", got, want)
	}
}

func TestRemoveProtocol(t *testing.T) {
	for in, want := range map[string]string{
		"http://example.com":   "example.com",
		"https://example.com/": "example.com",
		"example.com":          "example.com",
	} {
		if got := RemoveProtocol(in); got != want {
			t.Errorf("RemoveProtocol(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDomain(t *testing.T) {
	for in, want := range map[string]string{
		"https://www.Example.com/a/b?c=d#e": "example.com",
		"http://a.b.example.co:8080/x":      "example.co",
		"example.com?q=1":                   "example.com",
		"example.com#top":                   "example.com",
		"localhost":                         "localhost",
		"":                                  "",
	} {
		if got := Domain(in); got != want {
			t.Errorf("Domain(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestQuery(t *testing.T) {
	tests := map[string]map[string]string{
		"https://example.com":                         {},
		"https://example.com?a=1&b=x%20y&flag&c=d=e":  {"a": "1", "b": "x y", "c=d": "e"},
		"https://bing.com/news?url=https%3a%2f%2fx.y": {"url": "https://x.y"},
	}
	for in, want := range tests {
		if got := Query(in); !maps.Equal(got, want) {
			t.Errorf("Query(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestPath(t *testing.T) {
	for in, want := range map[string]string{
		"https://example.com/a/b?c=d": "a/b",
		"https://example.com":         "",
		"example.com/x/":              "x",
	} {
		if got := Path(in); got != want {
			t.Errorf("Path(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUUID(t *testing.T) {
	if a, b := UUID("https://a.com"), UUID("https://a.com"); a != b || len(a) != 36 {
		t.Errorf("UUID() = %q, %q; want a stable uuid", a, b)
	}
	if UUID("https://a.com") == UUID("https://b.com") {
		t.Error("different urls should have different uuids")
	}
}

func TestSecure(t *testing.T) {
	if !Secure("https://a.com") || Secure("http://a.com") || Secure("a.com") {
		t.Error("Secure() should only accept https urls")
	}
}
