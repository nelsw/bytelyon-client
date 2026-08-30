package api

import (
	"bytelyon-client/internal/provider/http"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

var (
	uri string
	hdr map[string][]string
)

func Init(url, key string) {
	if key == "" {
		log.Fatal().Msg("no api key provided")
	}
	uri = url + "/api"
	hdr = map[string][]string{
		"Authorization": {"Bearer " + key},
	}
}

func url(parts []any) string {
	var out strings.Builder
	out.WriteString(uri)
	for _, part := range parts {
		_, _ = fmt.Fprintf(&out, "/%v", part)
	}
	return out.String()
}

func Get(paths ...any) []byte {
	out, _ := http.Get(url(paths), hdr)
	return out
}

func Put(a any, paths ...any) int {
	out, err := http.Put(url(paths), a, hdr)
	if err != nil {
		log.Err(err).Any("a", a).Msg("failed to PUT")
		return -1
	}
	var m map[string]int
	if err = json.Unmarshal(out, &m); err != nil {
		log.Err(err).Msg("failed to unmarshal PUT response")
		return -1
	}
	if v, ok := m["id"]; ok {
		return v
	}
	return -1
}

func Post(img []byte, fileName, fieldName string, paths ...any) {
	if err := http.PostFile(url(paths), fileName, fieldName, img, hdr); err != nil {
		log.Err(err).Msg("failed to POST")
	}
}
