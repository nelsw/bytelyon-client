package news

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/http"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/nelsw/bytelyon-client/pkg/s3"
	"github.com/rs/zerolog/log"
)

func Fetch(m *bot.Model) {

	var arr []*Article

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

	if err := play.News(m.Headless, s.Keys()); err != nil {
		log.Err(err).Send()
	}

	path := filepath.Join(".storage", "sitemap", strconv.Itoa(m.ID))

	for _, a := range arr {
		p := filepath.Join(path, uuid.NewSHA1(uuid.NameSpaceURL, []byte(a.URL)).String())

		from := p + ".html"
		to := strings.ReplaceAll(from, ".storage/", "")
		_ = s3.Move(from, to)
		//_ = os.Remove(from)

		from = p + ".png"
		to = strings.ReplaceAll(from, ".storage/", "")
		_ = s3.Move(from, to)
		//_ = os.Remove(from)

		from = p + ".json"
		bytes, err := os.ReadFile(from)
		if err != nil {
			log.Warn().Err(err).Msg("failed to read page")
			return
		}
		//_ = os.Remove(from)

		var d model.Data[string, any]
		if err = json.Unmarshal(bytes, &d); err != nil {
			log.Warn().Err(err).Msg("failed to unmarshal page")
			return
		}

		d.Put("bot_id", m.ID)
		d.Put("source", a.Source)
		d.Put("title", a.Title)
		d.Put("publisher", a.Publisher)
		if a.Desc != "" {
			d.Put("description", a.Desc)
		}

		Save(m.ID, a)
	}
}

func save() {

}
