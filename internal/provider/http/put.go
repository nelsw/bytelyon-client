package http

import "net/http"

func Put(url string, data any, header http.Header) ([]byte, error) {
	out, _, err := do(http.MethodPut, url, data, header)
	return out, err
}
