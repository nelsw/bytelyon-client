package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nelsw/bytelyon-client/config"
	"github.com/nelsw/bytelyon-client/internal/news"
	"github.com/nelsw/bytelyon-client/internal/search"
	"github.com/nelsw/bytelyon-client/internal/sitemap"
	"github.com/nelsw/bytelyon-client/pkg/http"
	"github.com/nelsw/bytelyon-client/pkg/model"

	"github.com/rs/zerolog/log"
)

const maxLessConcurrency = 5
const maxFullConcurrency = 5

var (
	poll *time.Ticker
	wg   sync.WaitGroup
	full = make(chan *model.Bot)
	less = make(chan *model.Bot)
)

func init() {
	worker(maxFullConcurrency, full)
	worker(maxLessConcurrency, less)
	poll = time.NewTicker(15 * time.Second)
}

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-poll.C:
			handlePoll()
		case <-quit:
			handleQuit()
		}
	}
}

func handleQuit() {
	fmt.Println() // newline after Ctrl+C
	log.Log().Msg("quitting...")
	poll.Stop()
	close(less)
	close(full)
	wg.Wait()
	log.Log().Msg("👋")
	os.Exit(0)
}

func handlePoll() {
	log.Log().Msg("polling bots...")

	var bots []*model.Bot
	url := fmt.Sprintf("%s/api/bots", config.ApiHost())
	if out, err := http.Get(url, config.AuthHeader()); err != nil {
		log.Err(err).Msg("failed to get bots")
	} else if err = json.Unmarshal(out, &bots); err != nil {
		log.Err(err).Msg("failed to unmarshal bots")
	} else {
		log.Info().Int("size", len(bots)).Msg("got bots")
	}

	for _, b := range bots {
		if b.Headless {
			less <- b
		} else {
			full <- b
		}
	}
}

func worker(i int, ch chan *model.Bot) {

	ƒ := func(b *model.Bot) func() {
		return func() {
			log.Log().EmbedObject(b).Msg("working ...")
			defer func() {
				log.Log().Msg("putting bot...")
				url := fmt.Sprintf("%s/api/bots/%d", config.ApiHost(), b.ID)
				if _, err := http.Put(url, b, config.AuthHeader()); err != nil {
					log.Err(err).Msg("failed to put bot")
				}
				log.Log().EmbedObject(b).Msg("done")
			}()
			switch b.Type {
			case model.NewsBot:
				news.Handle(b.ID, b.Headless, b.Query, b.PlayedAt, b.Blacklist)
			case model.SearchBot:
				search.Handle(b.ChildID, b.ID, b.Headless, b.Query, b.Blacklist)
			case model.SitemapBot:
				sitemap.Handle(b.ChildID, b.ID, b.Headless, b.Query)
			}
		}
	}

	for range i {
		for b := range ch {
			wg.Go(ƒ(b))
		}
	}
}
