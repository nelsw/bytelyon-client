package url

import (
	"net/url"
	"regexp"
	"strings"
)

type Values url.Values

func (v Values) Encode() []byte {
	return []byte(url.Values(v).Encode())
}

var (
	browserFunction = regexp.MustCompile(`^(mailto|tel|sms|fax|callto|geo|javascript|about):.*`)
)

func removeProtocol(url string) string {
	url = strings.TrimPrefix(Clean(url), "http://")
	url = strings.TrimPrefix(url, "https://")
	return url
}

// Clean normalizes a URL string by trimming whitespace, converting to lowercase, and removing a trailing slash.
func Clean(url string) string {
	// trim whitespace jic
	url = strings.TrimSpace(url)
	// lowercase to normalize
	url = strings.ToLower(url)
	// remove trailing slash
	return strings.TrimSuffix(url, "/")
}

// Domain returns the domain name from a URL in lowercase.
// Unlink url.Parse, this ƒ does not require a protocol to determine a hostname.
func Domain(url string) string {

	url = removeProtocol(url)

	// remove path
	url = strings.Split(url, "/")[0]

	// remove query
	url = strings.Split(url, "?")[0]

	// remove fragment
	url = strings.Split(url, "#")[0]

	// remove port
	url = strings.Split(url, ":")[0]

	// remove subdomains
	for strings.Count(url, ".") > 1 {
		ss := strings.Split(url, ".")
		url = ss[len(ss)-2] + "." + ss[len(ss)-1]
	}

	return url
}

func Query(url string) map[string]string {

	var m = make(map[string]string)

	_, s, exists := strings.Cut(url, "?")
	if !exists {
		return m
	}

	for q := range strings.SplitSeq(s, "&") {
		if k, v, ok := strings.CutLast(q, "="); ok {
			m[k] = v
			continue
		}

	}
	return m
}

func Path(url string) (s string) {
	url = removeProtocol(url)
	_, s, _ = strings.Cut(url, "/")
	s, _, _ = strings.Cut(s, "?")
	return
}

func IsBrowserFunction(s string) bool {
	return browserFunction.MatchString(s)
}
