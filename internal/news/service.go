package news

import (
	"encoding/json"
	"os"
	"strings"
	"sync"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/http"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/scrape"
	"github.com/rs/zerolog/log"
)

func Build(m bot.Model) {

	q := model.NewQueue[*Article]()
	var g sync.WaitGroup
	c := make(chan *Article)

	ƒ := func(s Source) {

		rss, err := http.New(s.URL(m.Query)).Get().XML[RSS]()
		if err != nil {
			log.Err(err).Send()
			return
		}

		var wg sync.WaitGroup
		for _, a := range rss.Channel.Articles {
			wg.Go(func() {

				if !m.RanBefore(a.PublishedAt()) || !m.Blacklist.OK(a.Words()) {
					return
				}

				if a.IsGoogleNews() {
					a.URL = decodeGoogleLink(a.Link)
					a.Source = "Google News"
					if l, r, ok := strings.Cut(a.Title, " - "); ok {
						a.Publisher = r
						a.Title = l
						a.Desc = ""
					}
				} else {
					a.URL = decodeBingLink(a.Link)
					a.Publisher = a.Source
					a.Source = "Bing News"
				}

				g.Add(1)
				q.Push(a)
			})
		}
		wg.Wait()
	}

	var wg sync.WaitGroup
	wg.Go(func() { ƒ(BingNews) })
	wg.Go(func() { ƒ(GoogleNews) })
	wg.Wait()

	log.Info().Int("size", q.Len()).Msg("articles")

	tasks := make(chan *Article)
	for range maxAsync {
		wg.Go(func() {
			for a := range tasks {

				key, err := scrape.Page(m.Headless, a.URL, "news", m.ID)
				if err != nil {
					log.Warn().Err(err).Msg("failed to scrape page")
					continue
				}

				var bytes []byte
				if bytes, err = os.ReadFile(key + ".json"); err != nil {
					log.Warn().Err(err).Msg("failed to read page")
					continue
				}

				var d model.Data
				if err = json.Unmarshal(bytes, &d); err != nil {
					log.Warn().Err(err).Msg("failed to unmarshal page")
					continue
				}

				d.Put("bot_id", m.ID)
				d.Put("source", a.Source)
				d.Put("title", a.Title)
				d.Put("publisher", a.Publisher)
				if a.Desc != "" {
					d.Put("description", a.Desc)
				}

				Save(d)
			}
		})
	}

	go func() {
		for {
			if t, ok := q.Pop(); ok {
				tasks <- t
			} else {
				break
			}
		}
		close(c)
	}()

	go func() {
		q.Close()
	}()

	g.Wait()

	bot.Update(m.ID)
}
