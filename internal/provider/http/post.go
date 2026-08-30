package http

import (
	"bytelyon-client/internal/util/url"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"

func PostFile(u, fieldName, filename string, b []byte, h map[string][]string) error {

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	defer func() {
		_ = bodyWriter.Close()
	}()

	if fileWriter, err := bodyWriter.CreateFormFile(fieldName, filename); err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	} else if _, err = io.Copy(fileWriter, bytes.NewReader(b)); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	} else if err = bodyWriter.Close(); err != nil {
		return fmt.Errorf("failed to close body writer: %w", err)
	}

	if h == nil {
		h = map[string][]string{}
	}
	h["Content-Type"] = []string{bodyWriter.FormDataContentType()}

	_, _, err := do(http.MethodPost, u, bodyBuf, h)
	return err
}

func PostForm(u string, v url.Values) ([]byte, error) {
	return post(u, v.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
		"User-Agent":   userAgent,
	})
}

func PostJSON(u string, b []byte, h map[string][]string) ([]byte, error) {
	out, _, err := do(http.MethodPost, u, b, h)
	return out, err
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
