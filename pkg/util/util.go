package util

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"

	"github.com/rs/zerolog/log"
)

func Data(s string) []byte {
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
