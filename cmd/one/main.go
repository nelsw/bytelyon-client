package main

import (
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

	rob := bot.FindOne()
	if rob == nil {
		log.Info().Msg("No robots found")
		return
	}

	log.Info().EmbedObject(rob).Send()
	defer func() {
		bot.Update(rob.ID)
		postgres.Close()
		redis.Close()
	}()

	switch rob.Type {
	case bot.NewsType:
		news.Fetch(rob)
	case bot.SearchType:
		search.Build(rob)
	case bot.SitemapType:
		sitemap.Fetch(rob)
	}
}
