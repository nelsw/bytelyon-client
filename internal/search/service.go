package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/nelsw/bytelyon-client/pkg/store"
	"github.com/nelsw/bytelyon-client/pkg/uuid"
	"github.com/rs/zerolog/log"
)

func blocked(p playwright.Page) bool {
	return strings.HasPrefix(p.URL(), "https://www.google.com/sorry")
}

func handleCaptcha(p playwright.Page) bool {
	for {
		log.Warn().
			Str("url", p.URL()).
			//Int("seconds", 90-(3*i)).
			Msg("user must solve captcha...")

		if time.Sleep(3 * time.Second); !blocked(p) {
			log.Info().Msg("captcha solved - well done!")
			return true
		}
	}
	//log.Warn().
	//	Str("url", p.URL()).
	//	Msg("failed to solve captcha in time!")
	//return false
}

func Fetch(r *Result, headless bool) error {

	context, err := play.New(headless)
	if err != nil {
		return err
	}
	defer play.Close(context)

	var p playwright.Page
	if p, err = play.NewPage(context); err != nil {
		return err
	} else if _, err = play.GoTo(p, "https://www.google.com"); err != nil {
		return err
	} else if err = play.Click(p, "textarea[name='q']", 20, 50); err != nil {
		return err
	} else if err = play.Type(p, r.Query, 10, 40); err != nil {
		return err
	}

	play.Sleep(p, 100, 300)

	if err = p.Keyboard().Press("Enter"); err != nil {
		return err
	} else if err = play.Wait(p, playwright.LoadStateNetworkidle); err != nil {
		return err
	}

	if blocked(p) && !handleCaptcha(p) {
		return fmt.Errorf("blocked")
	}

	log.Info().Str("q", r.Query).Msg("reached SERP")

	if err = play.Wait(p, playwright.LoadStateDomcontentloaded); err != nil {
		return err
	}

	play.Sleep(p, 1_000, 2_500)

	if err = play.Scroll(p); err != nil {
		return err
	}

	play.IMG(p, r.IMG)
	fetchSimilarQueries(r, p)
	fetchOrganicProducts(r, context, p)
	fetchSponsoredProducts(r, context, p)
	fetchSponsoredResults(r, context, p)
	fetchOrganicResults(r, context, p)
	if err = store.Save(r, "search", r.BotID, uuid.FromURL(r.url).String()+".json"); err != nil {
		log.Err(err).Msg("failed to save search result")
	}

	return nil
}

func fetchSimilarQueries(r *Result, serp playwright.Page) {

	m := make(map[string]bool)
	var s string

	all, err := serp.Locator("div[data-notify-expansion]").All()
	log.Err(err).
		Stringer("section", SimilarQueries).
		Str("subsection", "exp").
		Int("count", len(all)).
		Send()
	if err == nil {
		for _, l := range all {
			s, err = l.GetAttribute("data-q")
			m[s] = err == nil
			log.Err(err).
				Stringer("section", SimilarQueries).
				Str("subsection", "exp").
				Str("text", s).
				Send()
		}
	}

	all, err = serp.Locator("div#botstuff").Locator("a").All()
	log.Err(err).
		Stringer("section", SimilarQueries).
		Str("subsection", "bot").
		Int("count", len(all)).
		Send()
	if err == nil {
		for _, l := range all {
			s, err = l.TextContent()
			m[s] = err == nil
			log.Err(err).
				Stringer("section", SimilarQueries).
				Str("subsection", "bot").
				Str("text", s).
				Send()
		}
	}

	for k, v := range m {
		if v && len(k) > 4 {
			r.Add(SimilarQueries, k)
		}
	}
}

func fetchSponsoredProducts(r *Result, x playwright.BrowserContext, serp playwright.Page) {

	all, err := serp.Locator("[data-dtld]").All()
	log.Err(err).Stringer("section", SponsoredProducts).Int("count", len(all)).Send()
	if err != nil {
		return
	}

	var p playwright.Page
	for _, l := range all {
		if p, err = play.NewTab(x, l); err == nil && p != nil {
			play.IMG(p, r.AddPage(SponsoredProducts, p.URL(), play.Title(p), play.HTML(p)).IMG)
			play.Sleep(p, 500, 1_000)
			_ = p.Close()
		}
	}
}

func fetchOrganicProducts(r *Result, x playwright.BrowserContext, serp playwright.Page) {

	all, err := serp.Locator(`product-viewer-entrypoint`).All()
	log.Err(err).Stringer("section", OrganicProducts).Int("count", len(all)).Send()
	if err != nil {
		return
	}

	var p playwright.Page
	for _, l := range all {

		// click the product image to display the product viewer elements
		if err = l.Locator("img").First().Click(playwright.LocatorClickOptions{Force: new(true)}); err != nil {
			log.Err(err).Msg("failed to click on product image")
			continue
		}

		// click the anchor nested inside the redirect element
		if p, err = play.NewTab(x, serp.Locator(`div[data-redirect-url] a`).First()); err == nil && p != nil {
			play.IMG(p, r.AddPage(OrganicProducts, p.URL(), play.Title(p), play.HTML(p)).IMG)
			play.Sleep(p, 500, 1_000)
			_ = p.Close()
		}
	}
}

func fetchSponsoredResults(r *Result, x playwright.BrowserContext, serp playwright.Page) {

	all, err := serp.Locator("[data-pcu]").All()
	log.Err(err).Stringer("section", SponsoredResults).Int("count", len(all)).Send()
	if err != nil {
		return
	}

	var p playwright.Page
	for _, l := range all {
		if p, err = play.NewTab(x, l); err == nil && p != nil {
			play.IMG(p, r.AddPage(SponsoredResults, p.URL(), play.Title(p), play.HTML(p)).IMG)
			play.Sleep(p, 500, 1_000)
			_ = p.Close()
		}
	}
}

func fetchOrganicResults(r *Result, x playwright.BrowserContext, serp playwright.Page) {

	all, err := serp.Locator("h3[id]").All()
	log.Err(err).Stringer("section", OrganicResults).Int("count", len(all)).Send()
	if err != nil {
		return
	}

	var p playwright.Page
	for _, l := range all {
		if p, err = play.NewTab(x, l.Locator("xpath=ancestor::a[1]")); err == nil && p != nil {
			play.IMG(p, r.AddPage(OrganicResults, p.URL(), play.Title(p), play.HTML(p)).IMG)
			play.Sleep(p, 500, 1_000)
			_ = p.Close()
		}
	}
}
