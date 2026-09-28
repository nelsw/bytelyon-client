package model

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"

	"github.com/rs/zerolog/log"
)

type Page struct {
	Error       string         `json:"error"`
	URL         string         `json:"url"`
	Title       string         `json:"title"`
	Meta        map[string]any `json:"meta"`
	Links       []string       `json:"links"`
	Body        string         `json:"body"`
	ImgSrc      string         `json:"img_src"`
	ImgAlt      string         `json:"img_alt"`
	Description string         `json:"description"`
	Keywords    []string       `json:"keywords"`
	Screenshot  string         `json:"screenshot"`
	Content     string         `json:"content"`
	Data        map[string]any `json:"data"`
}

func (p *Page) ScreenshotBytes() []byte {
	return p.bytes(p.Screenshot)
}

func (p *Page) ContentBytes() []byte {
	return p.bytes(p.Content)
}

func (p *Page) bytes(s string) []byte {
	compressed, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Err(err).Msg("Failed to decode string")
		return nil
	}

	var reader *gzip.Reader
	if reader, err = gzip.NewReader(bytes.NewReader(compressed)); err != nil {
		log.Err(err).Msg("Failed to create gzip reader")
		return nil
	}
	defer func() { _ = reader.Close() }()

	var out []byte
	if out, err = io.ReadAll(reader); err != nil {
		log.Err(err).Msg("Failed to read all uncompressed bytes")
		return nil
	}
	return out
}
