package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nelsw/bytelyon-client/internal/app"
	"github.com/nelsw/bytelyon-client/internal/bot"
)

func main() {

	app.Init(true)
	defer app.Close()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	tick := time.NewTicker(time.Minute * 15)

	app.HandleBots(bot.FindAll())
	for {
		select {
		case <-tick.C:
			app.HandleBots(bot.FindAll())
		case <-quit:
			tick.Stop()
			app.Close()
			return
		}
	}
}
