package job

import (
	"bytelyon-client/internal/job/news"
	"bytelyon-client/internal/model"
	"bytelyon-client/internal/provider/logs"
	"bytelyon-client/internal/provider/play"

	"github.com/mxschmitt/playwright-go"
	"golang.org/x/sync/errgroup"
)

func Do(b *model.Bot) {
	logs.Init("trace")
	defer b.Save()
	switch b.Type {
	case model.NewsBot:
		doNews(b)
	case model.SearchBot:
		doSearch(b)
	case model.SitemapBot:
		doSitemap(b)
	}
}

func doNews(b *model.Bot) {

	c := make(chan *model.Article)

	var g errgroup.Group
	g.SetLimit(5)
	ƒ := func(ctx playwright.BrowserContext) {
		for a := range c {
			g.Go(func() error {
				if content, err := play.ScrapeContent(ctx, a.URL); err == nil {
					a.Fill(content)
				}
				a.Save()
				return nil
			})
		}
	}

	go news.DoGoogle(c, b)
	go news.DoBing(c, b)

	play.It(b.Headless, ƒ)

	_ = g.Wait()
}

func doSearch(b *model.Bot) {
}

func doSitemap(b *model.Bot) {
}
