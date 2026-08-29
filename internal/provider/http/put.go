package http

import "net/http"

func Put(url string, data any, header http.Header) error {
	_, _, err := do(http.MethodPut, url, data, header)
	return err
}
