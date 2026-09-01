package news

import (
	"sync"
	"time"

	"github.com/nelsw/bytelyon-client/config"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/nelsw/bytelyon-client/pkg/store"
	"github.com/nelsw/bytelyon-client/pkg/uuid"
	"github.com/rs/zerolog/log"
)

const maxConcurrency = 5

func Handle(id int, headless bool, query string, after time.Time, ignore *model.Set[string]) error {

	context, err := play.New(headless)
	if err != nil {
		return err
	}
	defer play.Close(context)

	articles := Fetch(query, after, ignore)
	log.Info().Str("q", query).Msgf("articles found: %d", len(articles))
	if len(articles) == 0 {
		return nil
	}

	jobs := make(chan *Article)
	go func() {
		for _, a := range articles {
			log.Debug().EmbedObject(a).Msg("scraping article")
			jobs <- a
		}
		close(jobs)
	}()

	var wg sync.WaitGroup
	for range maxConcurrency {
		wg.Go(func() {
			var content string
			for a := range jobs {
				if content, err = play.ScrapeContent(context, a.URL); err == nil {
					a.Fill(content)
					log.Debug().EmbedObject(a).Msg("saving article")
					if save(id, a); !config.DryRun() {
						send(id, a)
					}
				}
			}
		})
	}
	wg.Wait()

	return nil
}

func save(botID int, a *Article) {
	name := uuid.FromURL(a.URL).String() + ".json"
	if err := store.Save(a, "news", botID, name); err != nil {
		log.Err(err).EmbedObject(a).Msg("failed to save article")
	}
}

func send(botID int, a *Article) {
	// todo - api put
}
