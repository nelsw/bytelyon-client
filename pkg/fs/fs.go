package fs

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
)

var (
	client *s3.Client
	once   sync.Once
	cfgErr error
)

// connect lazily creates the S3 client on first use.
func connect() (*s3.Client, error) {
	once.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			cfgErr = err
			return
		}
		client = s3.NewFromConfig(cfg)
	})
	return client, cfgErr
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

	var c *s3.Client
	if c, err = connect(); err != nil {
		return
	}

	_, err = c.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      new(os.Getenv("S3_BUCKET")),
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
