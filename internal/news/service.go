package news

import (
	"path/filepath"
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

func Fetch(
	botID int,
	query string,
	headless bool,
	lastRun time.Time,
	blacklist bot.Blacklist,
) {

	var arr []*Article

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

				arr = append(arr, a)
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

	if err := play.News(headless, s.Keys()); err != nil {
		log.Err(err).Send()
	}

	path := filepath.Join(".storage", "sitemap", strconv.Itoa(botID))

	for _, a := range arr {
		n := uuid.NewSHA1(uuid.NameSpaceURL, []byte(a.URL)).String()
		p := filepath.Join(path, n)
		_, _, d := play.HandleFiles(p)

		d.Put("bot_id", botID)
		d.Put("source", a.Source)
		d.Put("title", a.Title)
		d.Put("publisher", a.Publisher)
		if a.Desc != "" {
			d.Put("description", a.Desc)
		}

		UpsertArticle(botID, a)
	}
}
