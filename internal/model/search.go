package model

import (
	"bytelyon-client/internal/provider/api"
	"bytelyon-client/internal/provider/play"
	"bytelyon-client/internal/util/uuid"
	"fmt"
	"strings"

	"github.com/mxschmitt/playwright-go"
)

type Data map[string][]string
type Search struct {
	Bot   *Bot               `json:"-"`
	Pages map[string][]*Page `json:"-"`

	ID    int `json:"id"`
	Data  `json:"data"`
	Query string `json:"query"`
	BotID int    `json:"bot_id"`
	Screenshot
}

func NewSearch(b *Bot) *Search {
	u := "https://www.google.com?q=" + strings.ReplaceAll(b.Query, " ", "+")
	k := fmt.Sprintf("search/%d/%s/screenshot.png", b.ID, uuid.FromURL(u))
	return &Search{
		Bot:   b,
		BotID: b.ID,
		Query: b.Query,
		Pages: map[string][]*Page{
			"sponsored_products": {},
			"organic_results":    {},
			"sponsored_results":  {},
			"organic_products":   {},
		},
		Data: map[string][]string{
			"similar_queries": {},
		},
		ScreenshotKey: k,
	}
}

func (s *Search) Save() {
	s.Bot.ChildID = api.Put(s, "bots", s.Bot.ID, "searches")
	s.Screenshot.Save("serps", s.Bot.ChildID)
	for _, vv := range s.Pages {
		for _, v := range vv {
			v.Save()
		}
	}
}

func (s *Search) Do() {
	play.It(s.Bot.Headless, func(ctx playwright.BrowserContext) {
		serpPage, err := play.SearchGoogle(s.Query, ctx)
		if err != nil {
			return
		}
		s.ScreenshotData = play.Screenshot(serpPage)
		s.doSimilarQueries(serpPage)
		s.doSponsoredProducts(ctx, play.Locators(serpPage, "[data-dtld]"))
		s.doSponsoredResults(ctx, play.Locators(serpPage, "[data-pcu]"))
		s.doOrganicResults(ctx, play.Locators(serpPage, "h3[id]"))
		s.doOrganicProducts(ctx, serpPage)
		s.Save()
		_ = serpPage.Close()
	})
}

func (s *Search) doSponsoredProducts(x playwright.BrowserContext, ll []playwright.Locator) {
	var index int
	for _, l := range ll {
		if domain := play.Attribute(l, "data-dtld"); !s.Bot.Blacklist.OK(domain) {
			continue
		}
		merchantID := play.Attribute(l, "data-merchant-id")
		al := l.Locator(fmt.Sprintf("a[data-merchant-id=%s]", merchantID))
		href := play.Attribute(al, "href")
		src, img := play.Scrape(href, x)
		if doc, err := NewDoc(src); err == nil {
			p := NewPage(s.Bot, href, doc, img, index, "sponsored_products")
			s.Pages["sponsored_products"] = append(s.Pages["sponsored_products"], p)
			index++
		}
	}
}

func (s *Search) doOrganicProducts(x playwright.BrowserContext, serpPage playwright.Page) {
	var index int
	ll, err := serpPage.Locator("product-viewer-entrypoint").All()
	if err != nil {
		return
	}
	var doc *Doc
	for _, l := range ll {
		err = l.Locator("img").First().Click(playwright.LocatorClickOptions{
			Force:   new(true),
			Timeout: new(3000.0),
		})
		if err != nil {
			continue
		}
		sl := serpPage.Locator("div[data-redirect-url]").First()
		u := play.Attribute(sl, "data-redirect-url")
		src, img := play.Scrape(u, x)
		if doc, err = NewDoc(src); err == nil {
			p := NewPage(s.Bot, u, doc, img, index, "organic_products")
			s.Pages["organic_products"] = append(s.Pages["organic_products"], p)
			index++
		}
	}
}

func (s *Search) doSponsoredResults(x playwright.BrowserContext, ll []playwright.Locator) {
	var index int
	for _, l := range ll {
		href := play.Attribute(l, "href")
		src, img := play.Scrape(href, x)
		if doc, err := NewDoc(src); err == nil {
			p := NewPage(s.Bot, href, doc, img, index, "sponsored_results")
			s.Pages["sponsored_results"] = append(s.Pages["sponsored_results"], p)
			index++
		}
	}
}

func (s *Search) doOrganicResults(x playwright.BrowserContext, ll []playwright.Locator) {
	var index int
	for _, l := range ll {
		xl := l.Locator("xpath=ancestor::a[1]")
		href := play.Attribute(xl, "href")
		src, img := play.Scrape(href, x)
		if doc, err := NewDoc(src); err == nil {
			p := NewPage(s.Bot, href, doc, img, index, "organic_results")
			s.Pages["organic_results"] = append(s.Pages["organic_results"], p)
			index++
		}
	}
}

func (s *Search) doSimilarQueries(serpPage playwright.Page) {
	var texts []string
	for _, l := range play.Locators(serpPage, "div[data-notify-expansion]") {
		texts = append(texts, play.Attribute(l, "data-q"))
	}
	ll, err := serpPage.Locator("div#botstuff").Locator("a").All()
	if err == nil {
		var txt string
		for _, l := range ll {
			if txt, err = l.TextContent(); err == nil {
				texts = append(texts, txt)
			}
		}
	}
	for _, txt := range texts {
		if len(txt) > 4 {
			s.Data["similar_queries"] = append(s.Data["similar_queries"], txt)
		}
	}
}
