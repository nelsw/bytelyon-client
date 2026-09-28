package cache

import (
	"context"
	"errors"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nelsw/bytelyon-client/pkg/ssh"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	channel = "bots"
	pubSub  = 10
	pageDB  = 11
	linkDB  = 12
	xAnyDB  = 13
)

var (
	exp = time.Hour * 6
	ctx = context.Background()
	rcm = map[int]*redis.Client{
		pubSub: nil,
		pageDB: nil,
		linkDB: nil,
		xAnyDB: nil,
	}
)

func init() {
	for db, client := range rcm {
		if client == nil {
			rcm[db] = connect(db)
		}
	}
}

// connect lazily creates the client on first use.
func connect(db int) *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:         os.Getenv("REDIS_ADDR"),
		DB:           db,
		ReadTimeout:  -1,
		WriteTimeout: -1,
		Dialer:       ssh.DialFunc(),
	})

	if err := client.Ping(ctx).Err(); err != nil {
		log.Panic().Err(err).Str("addr", client.Options().Addr).Msg("redis ping failed")
	}
	return client
}

// Subscribe calls fn with each message published to the bots channel until the client is closed.
func Subscribe(fn func(payload string)) {
	sub := rcm[pubSub].Subscribe(ctx, channel)
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

func Put(db int, key string, val any) {
	if err := rcm[db].Set(ctx, key, val, exp).Err(); err != nil {
		log.Err(err).Send()
	}
}

func Get(db int, key string) (string, error) {
	return rcm[db].Get(ctx, key).Result()
}

func PutLink(key string, val any) {
	rcm[linkDB].Set(ctx, key, val, exp)
}

func GetLink(key string) (string, error) {
	return rcm[linkDB].Get(ctx, key).Result()
}

func PutPage(key string) {
	if err := rcm[pageDB].Set(ctx, key, time.Now().UnixMilli(), exp).Err(); err != nil {
		log.Err(err).Send()
	}
}

func GetPage(key string) time.Time {
	val, err := rcm[pageDB].Get(ctx, key).Int64()
	if err != nil {
		return time.Time{}
	}
	return time.UnixMilli(val)
}

func GetPageKeys(pattern string) []string {
	vals, err := rcm[pageDB].Keys(ctx, pattern).Result()
	if err != nil {
		log.Err(err).Send()
		return nil
	}
	return vals
}

func Close() {
	for _, client := range rcm {
		if client == nil {
			continue
		}
		if err := client.Close(); err != nil {
			log.Err(err).Send()
		}
	}
}
