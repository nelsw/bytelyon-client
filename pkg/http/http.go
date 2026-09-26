package http

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Response struct {
	Body []byte
	Err  error
	URL  string
}

func makeResponse(body io.ReadCloser, code int, url string) Response {
	if code >= 300 {
		return Response{Err: fmt.Errorf("HTTP error %d", code)}
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return Response{Err: err}
	}
	return Response{
		Body: b,
		URL:  url,
	}
}

func (r Response) XML[T any]() (t T, err error) {
	if err = r.Err; err == nil {
		err = xml.Unmarshal(r.Body, &t)
	}
	return
}

type Request struct {
	url string
	hdr http.Header
}

func New(URL string) *Request {
	return &Request{
		url: URL,
		hdr: http.Header{},
	}
}

func (r *Request) Header(key, value string) *Request {
	r.hdr.Add(key, value)
	return r
}

func (r *Request) Bearer(token string) *Request {
	return r.Header("Authorization", "Bearer "+token)
}

func (r *Request) Get() Response       { return r.do(http.MethodGet) }
func (r *Request) Put(a any) Response  { return r.do(http.MethodPut, a) }
func (r *Request) Post(a any) Response { return r.do(http.MethodPost, a) }

func (r *Request) do(method string, a ...any) Response {

	var buf io.Reader
	if len(a) > 0 {
		switch body := a[0].(type) {
		case string:
			buf = io.NopCloser(strings.NewReader(body))
		case []byte:
			buf = io.NopCloser(bytes.NewReader(body))
		default:
			b, err := json.Marshal(&body)
			if err != nil {
				return Response{Err: err}
			}
			buf = io.NopCloser(bytes.NewReader(b))
		}
	}

	req, err := http.NewRequest(method, r.url, buf)
	if err != nil {
		return Response{Err: err}
	}

	req.Header = r.hdr
	if req.Header.Get("Content-Type") == "" && (method == http.MethodPost || method == http.MethodPut) {
		req.Header.Set("Content-Type", "application/json")
	}

	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return Response{Err: err}
	}
	defer func() { _ = res.Body.Close() }()

	return makeResponse(res.Body, res.StatusCode, res.Request.URL.String())
}
