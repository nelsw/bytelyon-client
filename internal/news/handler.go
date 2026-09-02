package news

import (
	"sync"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/rs/zerolog/log"
)

const maxConcurrency = 5

func Handle(
	id int,
	h bool,
	q string,
	t time.Time,
	m map[string]bool,
) {

	arr := get(q, t, m)
	if len(arr) == 0 {
		return
	}

	context, err := play.New(h)
	if err != nil {
		return
	}
	defer play.Close(context)

	ch := make(chan *Article)
	for _, a := range arr {
		ch <- a
	}
	close(ch)

	var wg sync.WaitGroup
	for range maxConcurrency {
		wg.Go(func() {
			for a := range ch {
				var p playwright.Page
				log.Info().EmbedObject(a).Msg("scraping article")
				if p, err = play.NewPage(context); err != nil {
					log.Err(err).EmbedObject(a).Msg("failed to create page")
				} else if _, err = play.GoTo(p, a.Link); err != nil {
					log.Err(err).EmbedObject(a).Msg("failed to navigate to page")
				} else if err = play.Wait(p, playwright.LoadStateNetworkidle); err != nil {
					log.Err(err).EmbedObject(a).Msg("failed to wait for page to load")
				} else {
					play.Sleep(p, 500, 1000)
					a.URL = p.URL()
					a.Fill(play.HTML(p))
					_ = put(id, a)
				}
				if p != nil {
					_ = p.Close()
				}
			}
		})
	}
	wg.Wait()
}
