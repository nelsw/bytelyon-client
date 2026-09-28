package news

import (
	"encoding/xml"
	"strings"
	"sync"
	"time"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/http"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/rs/zerolog/log"
)

type rss struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		XMLName  xml.Name   `xml:"channel"`
		Articles []*Article `xml:"item"`
	} `xml:"channel"`
}

func Fetch(
	botID int,
	query string,
	headless bool,
	lastRun time.Time,
	blacklist bot.Blacklist,
) {

	ss := model.NewSyncSet[string]()

	ƒ := func(s Source) {

		out, err := http.New(s.URL(query)).Get().XML[rss]()
		if err != nil {
			log.Err(err).Send()
			return
		}

		var wg sync.WaitGroup
		for _, a := range out.Channel.Articles {
			wg.Go(func() {

				if lastRun.After(a.PublishedAt()) || !blacklist.OK(a.Words()) {
					return
				}

				switch s {
				case GoogleNews:
					a.URL = decodeGoogleLink(a.Link)
					a.Desc = ""
					if l, r, ok := strings.Cut(a.Title, " - "); ok {
						a.Publisher = r
						a.Title = l
					}
				case BingNews:
					a.URL = decodeBingLink(a.Link)
					a.Publisher = a.Source
				}

				if ss.Add(a.URL) {
					a.Source = string(s)
					a.BotID = botID
					go play.It(&Job{
						headless,
						a,
					})
				}
			})
		}
		wg.Wait()
	}

	var wg sync.WaitGroup
	wg.Go(func() { ƒ(BingNews) })
	wg.Go(func() { ƒ(GoogleNews) })
	wg.Wait()
}
