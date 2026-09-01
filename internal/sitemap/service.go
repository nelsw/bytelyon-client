package sitemap

import (
	"sync"

	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/rs/zerolog/log"
)

// Fetch walks the domain from its root, fetching at most P pages at a time, and
// records every url it reaches. It returns once nothing is left to visit.
//
// Work moves in a cycle: the frontier feeds toCrawl, crawlers fetch and feed
// toScrape, scrapers parse and feed the frontier again. Only the frontier is
// unbounded, which is what keeps the cycle from deadlocking on itself.
func Fetch(s *Sitemap, headless bool) error {
	context, err := play.New(headless)
	if err != nil {
		return err
	}
	defer play.Close(context)

	// seed the frontier first so the crawl is never mistaken for finished below
	s.Push(&Crawl{url: "https://" + s.Domain, depth: s.D})

	// fetching is the slow half, so P is the count of browser pages open at once
	var crawlers sync.WaitGroup
	for range s.P {
		// Crawl fetches urls on their own browser page and hands the results to a scraper.
		crawlers.Go(func() {
			for c := range s.toCrawl {
				content, screenshot := play.Scrape(c.url, context)
				if content == "" {
					// play.Scrape logs the reason; the url stays in the set as unreached
					s.pending.Done()
					continue
				}
				s.toScrape <- &Scrape{Crawl: c, content: content, screenshot: screenshot}
			}
		})
	}

	// parsing is cpu bound and comparatively cheap, but it should not be the
	// bottleneck that leaves browser pages idle, so it gets the same width
	var scrapers sync.WaitGroup
	for range s.P {
		// Scrape parses a fetched page, marks it reached, and feeds the on-domain links it finds back to the frontier.
		scrapers.Go(func() {
			for x := range s.toScrape {

				doc, err := model.NewDoc(x.content)
				if err != nil {
					log.Warn().Err(err).Str("url", x.url).Msg("failed to parse document")
					s.pending.Done()
					continue
				}

				// save it!
				//go NewPage(s.Bot, x.url, doc, x.screenshot, 0, "").Save()

				s.Update(x.url, true)

				for _, href := range doc.HREFs() {
					if next, ok := s.Parse(href); ok {
						s.Push(&Crawl{url: next, depth: x.depth - 1})
					}
				}

				// every child is admitted before its parent is retired, so the count only
				// reaches zero once the last page has given up its links
				s.pending.Done()
			}
		})
	}

	go func() {
		for {
			c, ok := s.frontier.Pop()
			if !ok {
				break
			}
			s.toCrawl <- c
		}
		close(s.toCrawl)
	}()

	// with nothing in flight, no url can discover another one, so we are done
	go func() {
		s.pending.Wait()
		s.frontier.Close()
	}()

	crawlers.Wait()
	close(s.toScrape)
	scrapers.Wait()

	log.Info().
		Str("domain", s.Domain).
		Int("pages", len(s.Keys())).
		Msg("sitemap built")

	return nil
}
