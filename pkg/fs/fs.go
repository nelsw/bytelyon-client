package fs

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
)

var client *s3.Client
var bucket string

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	bucket = os.Getenv("S3_BUCKET")
	client = s3.NewFromConfig(cfg)
}

// Put creates a new object or replaces an old object with a new object.
func Put(key string, data []byte) (err error) {

	defer func() {
		if err == nil {
			return
		}
		log.Err(err).
			Str("ƒ", "put").
			Str("key", key).
			Int("body", len(data)).
			Send()
	}()

	if len(key) == 0 {
		return errors.New("cannot put object with empty key")
	} else if len(data) == 0 {
		return errors.New("cannot put object with empty data")
	}

	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: new(http.DetectContentType(data)),
	})

	return
}

func Move(from, to string) (err error) {

	var data []byte

	defer func() {
		if err == nil {
			return
		}
		log.Err(err).
			Str("ƒ", "move").
			Str("from", from).
			Str("to", to).
			Int("body", len(data)).
			Send()
	}()

	if data, err = os.ReadFile(from); err != nil {
		return
	} else if err = Put(to, data); err != nil {
		return
	}
	err = os.Remove(from)
	return
}
