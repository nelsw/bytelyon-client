package job

import (
	"bytelyon-client/internal/job/news"
	"bytelyon-client/internal/model"
	"bytelyon-client/internal/provider/play"
	"sync"

	"github.com/mxschmitt/playwright-go"
)

func Do(b *model.Bot) {
	defer b.Save()
	switch b.Type {
	case model.NewsBot:
		doNews(b)
	}
}

func doNews(b *model.Bot) {
	var wg sync.WaitGroup
	c := make(chan *model.Article)
	ƒ := func(ctx playwright.BrowserContext) {
		for range 5 {
			for a := range c {
				wg.Go(func() {
					if content, err := play.ScrapeContent(ctx, a.URL); err == nil {
						a.Fill(content)
					}
					a.Save()
				})
			}
		}
	}

	go news.DoGoogle(c, b)
	go news.DoBing(c, b)

	play.It(b.Headless, ƒ)

	wg.Wait()
}
