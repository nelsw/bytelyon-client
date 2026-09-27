package news

import (
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goforj/godump"
	"github.com/google/uuid"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/http"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/rs/zerolog/log"
)

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

	urls := s.Keys()

	chunkiness := 2

	var chunks [][]string
	if len(urls) <= chunkiness {
		chunks = append(chunks, urls)
	} else {
		for i := range chunkiness {
			chunks = append(chunks, urls[i*len(urls)/chunkiness:(i+1)*len(urls)/chunkiness])
		}
	}

	godump.DumpJSON(chunks)

	for _, chunk := range chunks {
		wg.Go(func() {
			if err := play.Pages(bot.NewsType, botID, headless, chunk); err != nil {
				log.Err(err).Msg("while scraping news urls")
			}
		})
	}
	wg.Wait()

	for _, a := range arr {
		wg.Go(func() {
			n := uuid.NewSHA1(uuid.NameSpaceURL, []byte(a.URL)).String()
			p := filepath.Join(".storage", string(bot.NewsType), strconv.Itoa(botID), n)
			_, _, d := play.HandleFiles(p)

			d.Put("bot_id", botID)
			d.Put("source", a.Source)
			d.Put("title", a.Title)
			d.Put("publisher", a.Publisher)
			if a.Desc != "" {
				d.Put("description", a.Desc)
			}

			UpsertArticle(botID, a)
		})
	}
	wg.Wait()
}
