package sitemap

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/nelsw/bytelyon-client/pkg/url"
	"github.com/rs/zerolog/log"
)

const chunkiness = 2

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
		[]string{"https://" + domain},
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
	urls []string,
) error {

	log.Debug().
		Int("depth", depth).
		Str("domain", domain).
		Int("urls", len(urls)).
		Msg("scraping sitemap")

	if depth < 0 || len(urls) == 0 {
		return nil
	}

	chunks := make(chan []string)
	go func() {
		for chunk := range slices.Chunk(urls, 10) {
			chunks <- chunk
		}
		close(chunks)
	}()

	var wg sync.WaitGroup
	for range chunkiness {
		wg.Go(func() {
			for chunk := range chunks {
				if err := play.Pages(bot.SitemapType, sitemapID, headless, chunk); err != nil {
					log.Err(err).Msg("while scraping sitemap urls")
				}
			}
		})
	}
	wg.Wait()

	todo := model.MakeSet[string]()
	for _, u := range urls {
		wg.Go(func() {
			n := uuid.NewSHA1(uuid.NameSpaceURL, []byte(u)).String()
			p := filepath.Join(".storage", string(bot.SitemapType), strconv.Itoa(sitemapID), n)

			_, imgKey, d := play.HandleFiles(p)

			title, _ := d.Get("title").(string)
			meta, _ := d.Get("meta").(map[string]any)

			UpsertPage(sitemapID, domain, u, title, imgKey, meta)

			done.Add(u)

			links, _ := d.Get("links").([]any)

			log.Debug().Int("links", len(links)).Str("url", u).Send()

			for _, link := range links {
				if str, _ := link.(string); url.Domain(str) == domain && !done.Has(str) {
					todo.Add(str)
				}
			}
		})
	}
	wg.Wait()

	nextUrls := todo.Keys()
	nextDepth := depth - 1

	log.Debug().
		Int("next_level", nextDepth).
		Int("next_urls", len(nextUrls)).
		Msg("scraped pages")

	if nextDepth < 0 {
		done.AddAll(nextUrls)
		return nil
	}

	return fetch(
		domain,
		headless,
		sitemapID,
		nextDepth,
		done,
		nextUrls,
	)
}
