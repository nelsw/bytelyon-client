package model

import (
	"bytelyon-client/internal/provider/http"
	"bytelyon-client/internal/provider/play"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type News struct {
	Bot      *Bot       `json:"-"`
	Articles []*Article `json:"-"`
}

func NewNews(b *Bot) *News {
	return &News{
		Bot: b,
	}
}

func (n *News) Do() {

	var wg sync.WaitGroup
	wg.Go(n.doGoogle)
	wg.Go(n.doBing)
	wg.Wait()

	log.Info().Int("articles", len(n.Articles)).Msg("news articles")

	var eg errgroup.Group
	eg.SetLimit(5)
	play.It(n.Bot.Headless, func(context playwright.BrowserContext) {
		for _, a := range n.Articles {
			eg.Go(func() error {
				if content, err := play.ScrapeContent(context, a.URL); err == nil {
					a.Fill(content)
				}
				a.Save()
				return nil
			})
		}
		_ = eg.Wait()
	})
}

func (n *News) doBing() {
	q := strings.ReplaceAll(n.Bot.Query, " ", "+")
	u := fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", q)

	for _, item := range http.GetXML[RSS[BingItem]](u).Channel.Items {
		a := &Article{
			Bot:         n.Bot,
			Title:       item.Title,
			URL:         string(item.Link),
			PublishedAt: time.Time(item.PubDate),
			Description: item.Description,
			Publisher:   item.Source,
			ImgURL:      item.Image,
			Source:      "Bing News",
		}

		if a.OK() {
			n.Articles = append(n.Articles, a)
		}
	}
}

func (n *News) doGoogle() {
	q := strings.ReplaceAll(n.Bot.Query, " ", "+")
	u := fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", q)
	for _, item := range http.GetXML[RSS[GoogleItem]](u).Channel.Items {
		a := &Article{
			Bot:         n.Bot,
			Title:       string(item.Title),
			URL:         string(item.Link),
			PublishedAt: time.Time(item.PubDate),
			Source:      "Google News",
		}
		if a.OK() {
			n.Articles = append(n.Articles, a)
		}
	}
}
