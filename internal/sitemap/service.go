package sitemap

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/nelsw/bytelyon-client/pkg/url"
	"github.com/rs/zerolog/log"
)

var maxDepth int

func init() {
	maxDepth, _ = strconv.Atoi(os.Getenv("SITEMAP_DEPTH"))
}

func Fetch(
	sitemapID int,
	botID int,
	domain string,
	headless bool,
) {

	if sitemapID == 0 {
		log.Warn().Int("bot", botID).Msg("no sitemap for bot")
		return
	}

	urls := model.NewSyncSet[string]()
	if err := fetch(
		domain,
		headless,
		sitemapID,
		maxDepth,
		urls,
		"https://"+domain,
	); err != nil {
		log.Warn().Err(err).Msg("failed to scrape sitemap")
	}

	UpdateSitemap(sitemapID, urls.Keys())
}

func fetch(
	domain string,
	headless bool,
	sitemapID int,
	depth int,
	done *model.SyncSet[string],
	urls ...string,
) error {

	log.Debug().
		Int("depth", depth).
		Str("domain", domain).
		Int("urls", len(urls)).
		Msg("scraping sitemap")

	if depth < 0 || len(urls) == 0 {
		return nil
	}

	if err := play.Pages(bot.SitemapType, sitemapID, headless, urls); err != nil {
		log.Warn().Err(err).Msg("failed to scrape pages")
		return err
	}

	todo := model.NewSyncSet[string]()

	path := filepath.Join(".storage", "sitemap", strconv.Itoa(sitemapID))

	for _, u := range urls {
		n := uuid.NewSHA1(uuid.NameSpaceURL, []byte(u)).String()
		p := filepath.Join(path, n)

		_, imgKey, d := play.HandleFiles(p)

		title, _ := d.Get("title").(string)
		meta, _ := d.Get("meta").(map[string]any)

		UpsertPage(sitemapID, domain, u, title, imgKey, meta)

		done.Add(u)

		links, _ := d.Get("links").([]any)
		for _, link := range links {
			str, _ := link.(string)
			if str != "" && url.Domain(str) == domain && !done.Has(str) && !todo.Has(str) {
				todo.Add(str)
			}
		}
	}

	if depth > 0 {
		return fetch(domain, headless, sitemapID, depth-1, done, todo.Keys()...)
	}

	done.AddAll(todo.Keys())
	return nil
}
