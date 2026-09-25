package sitemap

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/s3"
	"github.com/nelsw/bytelyon-client/pkg/scrape"
	"github.com/nelsw/bytelyon-client/pkg/url"
	"github.com/rs/zerolog/log"
)

func Build(b bot.Model) {

	m := From(b)

	var g sync.WaitGroup

	q := model.NewQueue[*model.Entry[string, int]](
		model.NewEntry("https://"+b.Query, maxDepth),
	)

	c := make(chan *model.Entry[string, int])

	for range maxAsync {
		g.Go(func() {
			for e := range c {
				if m.Has(e.Key) {
					continue
				}

				key, err := scrape.Page(b.Headless, e.Key, "sitemap", b.ID)
				if err != nil {
					log.Warn().Err(err).Msg("failed to scrape page")
					continue
				}

				from := key + ".png"
				to := strings.ReplaceAll(from, ".storage", "bots")
				_ = s3.Move(from, to)

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

				SavePage(m, d)
				m.Add(e.Key)
				g.Done()

				nextDepth := e.Val - 1
				if nextDepth < 0 {
					continue
				}

				for _, link := range d.GetSlice("links") {
					u := fmt.Sprintf("%v", link)
					if url.Domain(u) == b.Query && !m.Has(u) {
						q.Push(model.NewEntry(u, nextDepth))
						g.Add(1)
					}
				}
			}
		})
	}

	go func() {
		for {
			if t, ok := q.Pop(); ok {
				c <- t
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

	Save(m.ID, m.Keys())
	bot.Update(b.ID)
}
