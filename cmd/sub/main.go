package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nelsw/bytelyon-client/internal/app"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/cache"
	"github.com/rs/zerolog/log"
)

func main() {
	app.Init()
	defer app.Close()

	cache.Subscribe(func(payload string) {
		var b bot.Model
		if err := json.Unmarshal([]byte(payload), &b); err != nil {
			log.Err(err).Msg("failed to unmarshal bot payload")
			return
		}
		app.HandleBot(b)
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.Close()
}
