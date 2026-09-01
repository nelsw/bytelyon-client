package sitemap

import (
	"cmp"
	"maps"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/nelsw/bytelyon-client/pkg/url"
)

const (
	defaultDepth    = 5
	defaultParallel = 25
)

type Sitemap struct {
	ID     int    `json:"id"`
	Domain string `json:"domain"`
	URL    string `json:"url"`
	*Set[string]

	D int // link depth to follow
	P int // pages to crawl in parallel

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

func NewResult(id int, domain string) *Sitemap {
	return &Sitemap{
		ID:       id,
		Domain:   domain,
		URL:      "https://" + domain,
		Set:      NewSet[string](),
		D:        defaultDepth,
		P:        defaultParallel,
		frontier: NewQueue[*Crawl](),
		toCrawl:  make(chan *Crawl),
		toScrape: make(chan *Scrape),
	}
}

//func (s *Sitemap) Save() {
//	a := map[string][]string{
//		"urls": s.Keys(),
//	}
//	api.Put(a, "bots", s.Bot.ID, "sitemaps")
//}

// Push admits a url to the frontier, skipping any that is out of depth or already
// seen. Every admitted url is counted, so Build knows when the crawl has settled.
func (s *Sitemap) Push(c *Crawl) {
	if c.depth <= 0 || !s.Put(c.url, false) {
		return
	}
	s.pending.Add(1)
	s.frontier.Push(c)
}

// Parse resolves an href to an absolute url on the domain, reporting whether it is worth crawling. Critical function.
func (s *Sitemap) Parse(href string) (string, bool) {

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
		(strings.HasPrefix(href, "https://") && url.Domain(href) != s.Domain) { // outbound
		return "", false
	}

	var u string
	switch {

	// the href is already absolute
	case strings.HasPrefix(href, "https://"):
		u = href

	// the href is missing the protocol
	case strings.HasPrefix(href, s.Domain):
		u = "https://" + href

	// the href is relative to the root url
	case strings.HasPrefix(href, "/"):
		u = "https://" + s.Domain + href

	// else the href is relative to the url it was found on
	default:
		u = "https://" + s.Domain + "/" + href
	}

	// files are not crawlable, and only the path can tell us it is one, since
	// every host ends in an extension of its own
	if path.Ext(url.Path(u)) != "" {
		return "", false
	}

	return u, true
}

type Set[K cmp.Ordered] struct {
	x sync.Mutex
	m map[K]bool
}

func NewSet[K cmp.Ordered]() *Set[K] {
	return &Set[K]{
		m: make(map[K]bool),
	}
}

func (s *Set[K]) Put(key K, val bool) bool {
	s.x.Lock()
	defer s.x.Unlock()
	if _, ok := s.m[key]; ok {
		return false
	}
	s.m[key] = val
	return true
}

// Update overwrites the value of an existing key, reporting whether it was present.
// Unlike Put, it does not add the key.
func (s *Set[K]) Update(key K, val bool) bool {
	s.x.Lock()
	defer s.x.Unlock()
	if _, ok := s.m[key]; !ok {
		return false
	}
	s.m[key] = val
	return true
}

func (s *Set[K]) Keys() []K {
	s.x.Lock()
	defer s.x.Unlock()
	return slices.Sorted(maps.Keys(maps.Clone(s.m)))
}

// Queue is an unbounded fifo. It buffers the crawl frontier
// so a producer never waits on a consumer that is, in turn,
// waiting on the producer.
type Queue[T any] struct {

	// mutex "locks" the queue for safe multithreaded access.
	mutex sync.Mutex

	// cond handles synchronization amongst asynchronous operations.
	cond sync.Cond

	// items are a generic collection of items.
	items []T

	// closed indicates if the queue is open or closed for bizness.
	closed bool
}

// NewQueue returns a new queue and initializes the necessary variables without the caller needing intimate knowledge.
func NewQueue[T any]() *Queue[T] {
	q := new(Queue[T])
	q.cond.L = &q.mutex
	return q
}

// Push appends an item and wakes a waiting Pop. It never blocks.
func (q *Queue[T]) Push(item T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, item)
	q.cond.Signal()
}

// Pop blocks until an item is available, reporting false once the queue has been closed and drained.
func (q *Queue[T]) Pop() (item T, ok bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// while the queue is empty and not closed
	for len(q.items) == 0 && !q.closed {
		q.cond.Wait()
	}

	// if the queue is closed and empty, return false
	if len(q.items) == 0 {
		return item, false
	}

	// return the first item and remove it from queue items
	item, q.items = q.items[0], q.items[1:]
	return item, true
}

// Close releases every waiting Pop once the remaining items have been drained.
func (q *Queue[T]) Close() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.closed = true
	q.cond.Broadcast()
}
