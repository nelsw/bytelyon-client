package model

import (
	"bytelyon-client/internal/provider/play"
	"bytelyon-client/internal/util/url"
	"path"
	"strings"
	"sync"

	"github.com/mxschmitt/playwright-go"
	"github.com/rs/zerolog/log"
)

const (
	defaultDepth    = 5
	defaultParallel = 25
)

type Sitemap struct {
	Bot    *Bot   `json:"-"`
	ID     int    `json:"-"`
	Domain string `json:"-"`
	U      string `json:"-"`

	*Set[string]

	D int // link depth to follow
	P int // pages to crawl in parallel
	X playwright.BrowserContext

	toCrawl  chan *Crawl  // urls waiting on a browser page
	toScrape chan *Scrape // fetched pages waiting on a parse

	frontier *Queue[*Crawl] // urls discovered but not yet handed to a crawler
	pending  sync.WaitGroup // urls admitted but not yet finished
}

// Crawl is a url waiting to be fetched.
type Crawl struct {
	url   string
	depth int
}

// Scrape is a fetched page waiting to be parsed.
type Scrape struct {
	*Crawl
	content    string
	screenshot []byte
}

type Option func(*Sitemap)

// WithBot identifies the parent bot of the sitemap.
func WithBot(b *Bot) Option { return func(s *Sitemap) { s.Bot = b } }

// WithID identifies the unique identifier of the sitemap.
func WithID(id int) Option { return func(s *Sitemap) { s.ID = id } }

// WithDepth limits how many links deep the crawl follows.
func WithDepth(d int) Option { return func(s *Sitemap) { s.D = d } }

// WithParallelism limits how many pages are crawled at once.
func WithParallelism(n int) Option { return func(s *Sitemap) { s.P = n } }

func NewSitemap(domain string, opts ...Option) *Sitemap {
	s := &Sitemap{
		Domain: domain,
		U:      "https://" + domain,
		Set:    NewSet[string](),
		D:      defaultDepth,
		P:      defaultParallel,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

//func (s *Sitemap) Save() { api.Put(s, "bots", s.BotID, "sitemaps") }

// Build walks the domain from its root, fetching at most P pages at a time, and
// records every url it reaches. It returns once nothing is left to visit.
//
// Work moves in a cycle: the frontier feeds toCrawl, crawlers fetch and feed
// toScrape, scrapers parse and feed the frontier again. Only the frontier is
// unbounded, which is what keeps the cycle from deadlocking on itself.
func (s *Sitemap) Build(x playwright.BrowserContext) {

	if s.P < 1 {
		s.P = 1
	}

	s.X = x
	s.frontier = NewQueue[*Crawl]()
	s.toCrawl = make(chan *Crawl)
	s.toScrape = make(chan *Scrape)

	// seed the frontier first so the crawl is never mistaken for finished below
	s.Push(&Crawl{url: s.U, depth: s.D})

	// fetching is the slow half, so P is the count of browser pages open at once
	var crawlers sync.WaitGroup
	for range s.P {
		crawlers.Go(s.Crawl)
	}

	// parsing is cpu bound and comparatively cheap, but it should not be the
	// bottleneck that leaves browser pages idle, so it gets the same width
	var scrapers sync.WaitGroup
	for range s.P {
		scrapers.Go(s.Scrape)
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
}

// Push admits a url to the frontier, skipping any that is out of depth or already
// seen. Every admitted url is counted, so Build knows when the crawl has settled.
func (s *Sitemap) Push(c *Crawl) {
	if c.depth <= 0 || !s.Put(c.url, false) {
		return
	}
	s.pending.Add(1)
	s.frontier.Push(c)
}

// Crawl fetches urls on their own browser page and hands the results to a scraper.
func (s *Sitemap) Crawl() {
	for c := range s.toCrawl {
		content, screenshot := play.Scrape(c.url, s.X)
		if content == "" {
			// play.Scrape logs the reason; the url stays in the set as unreached
			s.pending.Done()
			continue
		}
		s.toScrape <- &Scrape{Crawl: c, content: content, screenshot: screenshot}
	}
}

// Scrape parses a fetched page, marks it reached, and feeds the on-domain links it finds back to the frontier.
func (s *Sitemap) Scrape() {
	for x := range s.toScrape {

		doc, err := NewDoc(x.content)
		if err != nil {
			log.Warn().Err(err).Str("url", x.url).Msg("failed to parse document")
			s.pending.Done()
			continue
		}

		// save it!
		go NewSitemapPage(s.Bot, x.url, doc, x.screenshot).Save()

		s.Update(x.url, true)

		for _, href := range doc.HREFs() {
			if next, ok := s.Parse(s.Domain, href); ok {
				s.Push(&Crawl{url: next, depth: x.depth - 1})
			}
		}

		// every child is admitted before its parent is retired, so the count only
		// reaches zero once the last page has given up its links
		s.pending.Done()
	}
}

// Parse resolves an href to an absolute url on the domain, reporting whether it is worth crawling. Critical function.
func (s *Sitemap) Parse(domain, href string) (string, bool) {

	// trim whitespace, lowercase, and remove trailing slash
	href = url.Clean(href)

	// a fragment points back at the page it was found on, so drop it before dedupe
	href, _, _ = strings.Cut(href, "#")

	// protocol relative hrefs are https here; the domain checks below still apply
	if strings.HasPrefix(href, "//") {
		href = "https:" + href
	}

	// if the href is ...
	if href == "" || // empty, or the page it was found on
		url.IsBrowserFunction(href) || // browser function
		strings.HasPrefix(href, "http://") || // insecure
		(strings.HasPrefix(href, "https://") && url.Domain(href) != domain) { // outbound
		return "", false
	}

	var u string
	switch {

	// the href is already absolute
	case strings.HasPrefix(href, "https://"):
		u = href

	// the href is missing the protocol
	case strings.HasPrefix(href, domain):
		u = "https://" + href

	// the href is relative to the root url
	case strings.HasPrefix(href, "/"):
		u = "https://" + domain + href

	// else the href is relative to the url it was found on
	default:
		u = "https://" + domain + "/" + href
	}

	// files are not crawlable, and only the path can tell us it is one, since
	// every host ends in an extension of its own
	if path.Ext(url.Path(u)) != "" {
		return "", false
	}

	return u, true
}
