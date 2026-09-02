package http

import (
	"errors"
	"math/rand"
	"net/http"
	"time"
)

const (
	retryAttemptMax  = 3
	retryIntervalMax = 500 * time.Millisecond
)

func Get(url string, header http.Header) ([]byte, error) {

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
