package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/internal/news"
	"github.com/nelsw/bytelyon-client/internal/search"
	"github.com/nelsw/bytelyon-client/internal/sitemap"
	"github.com/nelsw/bytelyon-client/pkg/logs"
	"github.com/nelsw/bytelyon-client/pkg/postgres"
	"github.com/nelsw/bytelyon-client/pkg/redis"
	"github.com/rs/zerolog/log"
)

func init() {
	logs.Init()
	logs.Banner()
}

func main() {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Info().Msg("listening for quit signal (Ctrl+C)")

	handleAll()

	tick := time.NewTicker(time.Minute * 15)

	for {
		select {

		case <-tick.C:
			handleAll()
		case <-quit:
			tick.Stop()
			fmt.Println() // newline in buffer
			log.Info().Msg("received quit signal")

			redis.Close()
			postgres.Close()

			log.Info().Msg("goodbyte 👋")
			os.Exit(0)
		}
	}
}

func handleAll() {
	bots := bot.FindAll()
	log.Info().Msgf("found %d bots", len(bots))
	for _, b := range bots {
		switch b.Type {
		case bot.NewsType:
			news.Fetch(&b)
		case bot.SearchType:
			search.Build(&b)
		case bot.SitemapType:
			sitemap.Fetch(&b)
		}
		bot.Update(b.ID)
	}
}
