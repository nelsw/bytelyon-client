package cache

import (
	"context"
	"errors"
	"os"
	"strconv"
	"sync"
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
	mu     sync.Mutex
)

// connect lazily creates the client on first use.
func connect() *redis.Client {
	mu.Lock()
	defer mu.Unlock()

	if client != nil {
		return client
	}

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
	return client
}

// Subscribe calls fn with each message published to the bots channel until the client is closed.
func Subscribe(fn func(payload string)) {
	sub := connect().Subscribe(ctx, channel)
	defer func(sub *redis.PubSub) {
		if err := sub.Close(); err != nil && !errors.Is(err, redis.ErrClosed) {
			log.Err(err).Send()
		}
	}(sub)

	for {
		msg, err := sub.ReceiveMessage(ctx)
		if errors.Is(err, redis.ErrClosed) {
			return
		} else if err != nil {
			log.Warn().Err(err).Send()
			continue
		}
		fn(msg.Payload)
	}
}

func Put(key string, val any) {
	if err := connect().Set(ctx, key, val, 6*time.Hour).Err(); err != nil {
		log.Err(err).Send()
	}
}

func Get(key string) (string, error) {
	return connect().Get(ctx, key).Result()
}

func Close() {
	mu.Lock()
	defer mu.Unlock()

	if client == nil {
		return
	}
	if err := client.Close(); err != nil {
		log.Err(err).Send()
	}
	client = nil
}
