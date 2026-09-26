package app

import (
	"fmt"
	"sync/atomic"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/internal/news"
	"github.com/nelsw/bytelyon-client/internal/search"
	"github.com/nelsw/bytelyon-client/internal/sitemap"
	"github.com/nelsw/bytelyon-client/pkg/cache"
	"github.com/nelsw/bytelyon-client/pkg/db"
	"github.com/nelsw/bytelyon-client/pkg/logs"
	"github.com/nelsw/bytelyon-client/pkg/ssh"
	"github.com/rs/zerolog/log"
)

// working counts in-flight bot handlers so Close can wait for them to finish.
var working atomic.Int32

func Init() {
	logs.Init()
	log.Info().Msg("welcome 🦁")
}

func Close() {
	for working.Load() > 0 {
		time.Sleep(time.Second)
	}
	cache.Close()
	db.Close()
	ssh.Close()
	fmt.Println() // newline in buffer
	log.Info().Msg("goodbyte 👋")
}

func HandleBots(arr []bot.Model) {
	working.Add(1)
	defer working.Add(-1)
	for _, b := range arr {
		HandleBot(b)
	}
}

func HandleBot(b bot.Model) {

	working.Add(1)
	defer working.Add(-1)

	if b.ID == 0 {
		log.Info().Msg("bot not found")
		return
	}

	log.Info().EmbedObject(&b).Send()

	defer bot.UpdateFn(b)()

	switch b.Type {
	case bot.NewsType:
		news.Fetch(b.ID, b.Query, b.Headless, b.LastRun(), b.Blacklist)
	case bot.SearchType:
		search.Fetch(b.ChildID, b.Query, b.Headless)
	case bot.SitemapType:
		sitemap.Fetch(b.ChildID, b.ID, b.Query, b.Headless)
	}
}
