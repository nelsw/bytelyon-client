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

const channel = "bots"

const (
	pubSub = 10
	pages  = 11
	unk    = 15
)

var (
	ctx     = context.Background()
	clients = map[int]*redis.Client{
		pubSub: nil,
		pages:  nil,
		unk:    nil,
	}
)

func init() {
	for db, client := range clients {
		if client == nil {
			clients[db] = connect(db)
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
	sub := clients[pubSub].Subscribe(ctx, channel)
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
	if err := clients[unk].Set(ctx, key, val, 6*time.Hour).Err(); err != nil {
		log.Err(err).Send()
	}
}

func Get(key string) (string, error) {
	return clients[unk].Get(ctx, key).Result()
}

func PutTime(key string, val time.Time) {
	if err := clients[pages].Set(ctx, key, val.UnixMilli(), 6*time.Hour).Err(); err != nil {
		log.Err(err).Send()
	}
}

func GetTime(key string) time.Time {
	val, err := clients[pages].Get(ctx, key).Int64()
	if err != nil {
		return time.Time{}
	}
	return time.UnixMilli(val)
}

func Close() {
	for _, client := range clients {
		if client == nil {
			continue
		}
		if err := client.Close(); err != nil {
			log.Err(err).Send()
		}
	}
}
