package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nelsw/bytelyon-client/pkg/url"

	"github.com/rs/zerolog/log"
)

var client = http.Client{Timeout: 10 * time.Second}

func Put(url string, data any, header http.Header) ([]byte, error) {
	out, _, err := do(http.MethodPut, url, data, header)
	return out, err
}

func do(method, URL string, body any, header http.Header) (b []byte, code int, err error) {

	ctx := log.With().
		Str("ƒ", method).
		Str("domain", url.Domain(URL)).
		Str("path", url.Path(URL))
	if q := url.Query(URL); len(q) > 0 {
		ctx = ctx.Any("query", q)
	}
	l := ctx.Logger()

	l.Trace().Send()

	var buf io.Reader
	if body != nil {
		if v, ok := body.(io.Reader); ok {
			buf = v
		} else {
			if b, err = json.Marshal(&body); err != nil {
				l.Err(err).Any("body", body).Msg("failed to marshal body")
				return
			}
			buf = bytes.NewBuffer(b)
		}
	}

	var req *http.Request
	if req, err = http.NewRequest(method, URL, buf); err != nil {
		l.Err(err).Msg("failed to create request")
		return
	}

	if header != nil {
		req.Header = header
	}

	if req.Header.Get("Content-Type") == "" && (method == http.MethodPost || method == http.MethodPut) {
		req.Header.Set("Content-Type", "application/json")
	}

	var res *http.Response
	if res, err = client.Do(req); err != nil {
		l.Err(err).Msg("failed to do request")
		return
	}
	defer func() { _ = res.Body.Close() }()

	if b, err = io.ReadAll(res.Body); err != nil {
		l.Err(err).Msg("failed to read response")
	} else if code = res.StatusCode; code > 299 {
		err = fmt.Errorf("[%d] %s", res.StatusCode, string(b))
	}

	return
}
