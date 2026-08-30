package main

import (
	"bytelyon-client/internal/model"
	"bytelyon-client/internal/provider/api"
	"bytelyon-client/internal/provider/logs"
	"bytelyon-client/internal/provider/play"
	"flag"
	"maps"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

func init() {
	var lvl, url, key string
	flag.StringVar(&lvl, "log", "debug", "log level trace->disabled")
	flag.StringVar(&url, "url", "https://bytelyon.com", "web app api url")
	flag.StringVar(&key, "key", "", "client api key")
	flag.Parse()

	logs.Init(lvl)
	api.Init(url, key)
	play.Init()

	log.Log().Msg(`🦁`)
	log.Log().Msg(`🦁  ByteLyon`)
	log.Log().Msg(`🦁`)
}

func main() {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	todo := make(map[int]*model.Bot)

	ƒ := func(bot *model.Bot) {

		defer func() {
			bot.Save()
			delete(todo, bot.ID)
		}()

		log.Info().EmbedObject(bot).Send()

		switch bot.Type {
		case model.NewsBot:
			model.NewNews(bot).Do()
		case model.SearchBot:
			model.NewSearch(bot).Do()
		case model.SitemapBot:
			model.NewSitemap(bot).Do()
		}
	}

	poller := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-quit:
			poller.Stop()
			log.Log().Msgf("\n👋\n")
			return
		case <-poller.C:
			for _, bot := range model.Bots() {
				if _, ok := todo[bot.ID]; !ok {
					todo[bot.ID] = bot
				}
			}
		default:
			for key := range maps.Keys(todo) {
				ƒ(todo[key])
			}
		}
	}
}
