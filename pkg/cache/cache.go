package cache

import (
	"context"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nelsw/bytelyon-client/pkg/ssh"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const channel = "bots"

var (
	ctx    = context.Background()
	client *redis.Client
)

func init() {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	client = redis.NewClient(&redis.Options{
		Addr:         os.Getenv("REDIS_ADDR"),
		DB:           db,
		ReadTimeout:  -1,
		WriteTimeout: -1,
		Dialer:       ssh.DialFunc(),
	})

	if err := client.Ping(ctx).Err(); err != nil {
		log.Err(err).Str("addr", client.Options().Addr).Msg("redis ping failed")
	}
}

func Subscribe(fn func(payload string)) {
	sub := client.Subscribe(ctx, channel)
	defer func(sub *redis.PubSub) {
		if err := sub.Close(); err != nil {
			log.Err(err).Send()
		}
	}(sub)

	for {
		msg, err := sub.ReceiveMessage(ctx)
		if err != nil {
			log.Warn().Err(err).Send()
			continue
		}
		fn(msg.Payload)
	}
}

func Put(key string, val any) {
	if err := client.Set(ctx, key, val, 6*time.Hour).Err(); err != nil {
		log.Err(err).Send()
	}
}

func Get(key string) (string, error) {
	return client.Get(ctx, key).Result()
}

func Close() {
	if err := client.Close(); err != nil {
		log.Err(err).Send()
	}
}
