package api

import (
	"bytelyon-client/internal/provider/http"
	"fmt"
	"strings"
)

var (
	uri string
	hdr map[string][]string
)

func Init(url, key string) {
	uri = url + "/api"
	hdr = map[string][]string{
		"Authorization": {key},
	}
}

func url(parts ...any) string {
	var out strings.Builder
	out.WriteString(uri)
	for _, part := range parts {
		out.WriteString(fmt.Sprintf("/%v", part))
	}
	return out.String()
}

func Get(paths ...any) []byte {
	out, _ := http.Get(url(paths), hdr)
	return out
}

func Put(a any, paths ...any) {
	_ = http.Put(url(paths), a, hdr)
}
