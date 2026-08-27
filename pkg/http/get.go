package http

import (
	"bytes"
	"encoding/xml"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/net/html"
)

const (
	retryAttemptMax  = 3
	retryIntervalMax = 500 * time.Millisecond
)

func GetNode(url string) (node *html.Node, err error) {
	var out []byte
	if out, err = Get(url, nil); err == nil {
		node, err = html.Parse(bytes.NewReader(out))
	}
	return
}

func GetXML[T any](url string) (t T, err error) {
	var out []byte
	if out, err = Get(url, nil); err == nil {
		err = xml.Unmarshal(out, &t)
	}
	return
}

func Get(url string, opt ...http.Header) ([]byte, error) {

	retryInterval := time.Millisecond
	nextRetryInterval := func() time.Duration {
		// Add 10% jitter.
		interval := retryInterval + time.Duration(rand.Intn(int(retryInterval/10)))
		// Double and clamp for next time.
		retryInterval *= 2
		if retryInterval > retryIntervalMax {
			retryInterval = retryIntervalMax
		}
		return interval
	}

	var errs error
	var header http.Header
	if len(opt) > 0 {
		header = opt[0]
	}

	for range retryAttemptMax {
		if data, code, err := get(url, header); code == http.StatusOK {
			return data, nil
		} else if errs = errors.Join(errs, err); code != http.StatusTooManyRequests {
			break
		}
		time.Sleep(nextRetryInterval())
	}
	return nil, errs
}

func get(url string, header http.Header) ([]byte, int, error) {
	return do(http.MethodGet, url, nil, header)
}
