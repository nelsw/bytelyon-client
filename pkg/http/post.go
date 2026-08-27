package http

import (
	"bytelyon-client/pkg/url"
	"bytes"
	"io"
	"net/http"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"

func PostForm(u string, v url.Values) ([]byte, error) {
	return post(u, v.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
		"User-Agent":   userAgent,
	})
}

func PostJSON(u string, b []byte, h map[string]string) ([]byte, error) {
	h["Content-Type"] = "application/json"
	return post(u, b, h)
}

func post(URL string, b []byte, h map[string]string) ([]byte, error) {

	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	for k, v := range h {
		req.Header.Set(k, v)
	}

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	return io.ReadAll(res.Body)
}
