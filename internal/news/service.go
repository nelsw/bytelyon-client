package news

import (
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/http"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/rs/zerolog/log"
)

const chunkiness = 2

func Fetch(
	botID int,
	query string,
	headless bool,
	lastRun time.Time,
	blacklist bot.Blacklist,
) {

	var arr []*Article
	var mu sync.Mutex

	ƒ := func(s Source) {

		rss, err := http.New(s.URL(query)).Get().XML[RSS]()
		if err != nil {
			log.Err(err).Send()
			return
		}

		var wg sync.WaitGroup
		for _, a := range rss.Channel.Articles {
			wg.Go(func() {

				if lastRun.After(a.PublishedAt()) || !blacklist.OK(a.Words()) {
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

				mu.Lock()
				arr = append(arr, a)
				mu.Unlock()
			})
		}
		wg.Wait()
	}

	var wg sync.WaitGroup
	wg.Go(func() { ƒ(BingNews) })
	wg.Go(func() { ƒ(GoogleNews) })
	wg.Wait()

	s := model.MakeSet[string]()
	for _, a := range arr {
		s.Add(a.URL)
	}

	log.Info().Int("size", s.Len()).Msg("articles")
	if s.Len() == 0 {
		return
	}

	chunks := make(chan []string)
	go func() {
		for chunk := range slices.Chunk(s.Keys(), 10) {
			chunks <- chunk
		}
		close(chunks)
	}()

	for range chunkiness {
		wg.Go(func() {
			for chunk := range chunks {
				wg.Go(func() {
					if err := play.Pages(bot.NewsType, botID, headless, chunk); err != nil {
						log.Err(err).Msg("while scraping news urls")
					}
				})
			}
		})
	}
	wg.Wait()

	for _, a := range arr {
		wg.Go(func() {
			n := uuid.NewSHA1(uuid.NameSpaceURL, []byte(a.URL)).String()
			p := filepath.Join(".storage", string(bot.NewsType), strconv.Itoa(botID), n)
			_, _, d := play.HandleFiles(p)

			a.Body = d.Get("body").(string)
			a.Source = d.Get("source").(string)
			a.Title = d.Get("title").(string)
			a.Publisher = d.Get("publisher").(string)
			if a.Desc != "" {
				a.Desc = d.Get("description").(string)
			}

			UpsertArticle(botID, a)
		})
	}
	wg.Wait()
}
