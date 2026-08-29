package gzip

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/rs/zerolog/log"
)

func Decompress(in []byte) (out []byte, err error) {
	var reader *gzip.Reader
	if reader, err = gzip.NewReader(bytes.NewReader(in)); err != nil {
		log.Err(err).Msg("failed to initialize gzip reader")
		return
	}
	defer func() {
		_ = reader.Close()
	}()

	if out, err = io.ReadAll(reader); err != nil {
		log.Err(err).Msg("failed to read all uncompressed bytes")
	}
	return
}
