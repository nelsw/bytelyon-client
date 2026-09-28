package cache

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nelsw/bytelyon-client/pkg/ssh"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	subCh  = "bots"
	pubCh  = "evts"
	pubSub = 10
	store  = 11
)

var (
	exp = time.Hour * 6
	ctx = context.Background()
	rcm = map[int]*redis.Client{}
	mu  sync.Mutex
)

// connect lazily creates the client for db on first use and reuses it thereafter.
func connect(db int) *redis.Client {
	mu.Lock()
	defer mu.Unlock()

	if client, ok := rcm[db]; ok {
		return client
	}

	opt := redis.Options{
		Addr:         cmp.Or(os.Getenv("REDIS_ADDR"), "127.0.0.1:6379"),
		DB:           db,
		ReadTimeout:  -1,
		WriteTimeout: -1,
	}

	if os.Getenv("APP_ENV") == "prod" {
		opt.Dialer = ssh.DialFunc()
	}

	client := redis.NewClient(&opt)
	if err := client.Ping(ctx).Err(); err != nil {
		log.Err(err).Str("addr", client.Options().Addr).Msg("redis ping failed")
	}
	rcm[db] = client
	return client
}

// Subscribe calls fn with each message published to the bots channel until the client is closed.
func Subscribe(fn func(payload string)) {
	sub := connect(pubSub).Subscribe(ctx, subCh)
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

func Publish(botID int, message string) {
	connect(pubSub).Publish(ctx, pubCh, fmt.Sprintf(`{"id": %d, "message": "%s"}`, botID, message))
}

func Put(db int, key string, val any) {
	if err := connect(db).Set(ctx, key, val, exp).Err(); err != nil {
		log.Err(err).Send()
	}
}

func SetStr(key string, val any) {
	connect(store).Set(ctx, key, val, exp)
}

func GetStr(key string) (string, error) {
	return connect(store).Get(ctx, key).Result()
}

func SetTime(key string, t time.Time) {
	connect(store).Set(ctx, key, t.UnixMilli(), exp)
}

func GetTime(key string) time.Time {
	val, err := connect(store).Get(ctx, key).Int64()
	if err != nil {
		return time.Time{}
	}
	return time.UnixMilli(val)
}

func Keys(pattern string) []string {
	vals, err := connect(store).Keys(ctx, pattern).Result()
	if err != nil {
		log.Err(err).Send()
		return nil
	}
	return vals
}

func Del(key string) error {
	return connect(store).Del(ctx, key).Err()
}

func Decr(key string) (int64, error) {
	return connect(store).Decr(ctx, key).Result()
}

func Incr(key string) error {
	return connect(store).Incr(ctx, key).Err()
}

func Close() {
	mu.Lock()
	defer mu.Unlock()

	for db, client := range rcm {
		if err := client.Close(); err != nil {
			log.Err(err).Send()
		}
		delete(rcm, db)
	}
}
