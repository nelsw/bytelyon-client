package search

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mxschmitt/playwright-go"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/rs/zerolog/log"
)

func Fetch(r *Result, headless bool) error {

	context, err := play.New(headless)
	if err != nil {
		return err
	}
	defer play.Close(context)

	var serp playwright.Page
	if serp, err = play.SearchGoogle(r.Query, context); err != nil {
		return err
	}

	play.Screenshot(serp, r.ScreenshotKey)

	fetchOrganicProducts(r, context, serp)

	var wg sync.WaitGroup
	wg.Go(func() { fetchSimilarQueries(r, serp) })
	wg.Go(func() { fetchSponsoredProducts(r, context, serp) })
	wg.Go(func() { fetchSponsoredResults(r, context, serp) })
	wg.Go(func() { fetchOrganicResults(r, context, serp) })
	wg.Wait()

	return nil
}

func fetchSimilarQueries(r *Result, serpPage playwright.Page) {
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
			r.Add(SimilarQueries, txt)
		}
	}
}

func fetchSponsoredProducts(r *Result, x playwright.BrowserContext, serp playwright.Page) {
	for _, l := range play.Locators(serp, "[data-dtld]") {

		domain := play.Attribute(l, "data-dtld")
		if _, ok := r.Ignore[domain]; ok {
			continue
		}

		merchantID := play.Attribute(l, "data-merchant-id")
		al := l.Locator(fmt.Sprintf("a[data-merchant-id=%s]", merchantID))
		href := play.Attribute(al, "href")

		r.Add(SponsoredProducts, scrape(href, x))
	}
}

func fetchOrganicProducts(r *Result, x playwright.BrowserContext, serpPage playwright.Page) {
	for _, l := range play.Locators(serpPage, "product-viewer-entrypoint") {
		err := l.Locator("img").First().Click(playwright.LocatorClickOptions{
			Force:   new(true),
			Timeout: new(3000.0),
		})
		if err != nil {
			continue
		}
		sl := serpPage.Locator("div[data-redirect-url]").First()
		u := play.Attribute(sl, "data-redirect-url")
		r.Add(OrganicProducts, scrape(u, x))
	}
}

func fetchSponsoredResults(r *Result, x playwright.BrowserContext, serp playwright.Page) {
	for _, l := range play.Locators(serp, "[data-pcu]") {
		href := play.Attribute(l, "href")
		r.Add(SponsoredResults, scrape(href, x))
	}
}

func fetchOrganicResults(r *Result, x playwright.BrowserContext, serp playwright.Page) {
	for _, l := range play.Locators(serp, "h3[id]") {
		xl := l.Locator("xpath=ancestor::a[1]")
		href := play.Attribute(xl, "href")
		if strings.HasPrefix(href, "/") {
			href = serp.URL() + href
		}
		r.Add(OrganicResults, scrape(href, x))
	}
}

func scrape(url string, ctx playwright.BrowserContext) *Page {

	l := log.With().
		Str("ƒ", "scrape").
		Str("url", url).
		Logger()

	l.Trace().Send()

	page, err := play.NewPage(ctx)
	if err != nil {
		l.Warn().Msgf("scrape failed: %s", err.Error())
		return nil
	}

	defer func() { _ = page.Close() }()

	if err = play.Visit(page, url); err != nil {
		l.Warn().Msgf("Visit failed: %s", err.Error())
		return nil
	}

	l.Debug().Send()

	p := NewPage(
		page.URL(),
		play.Title(page),
		play.Content(page),
	)
	play.Screenshot(page, p.ScreenshotKey)
	return p
}
