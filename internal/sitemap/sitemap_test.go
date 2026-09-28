package sitemap

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nelsw/bytelyon-client/internal/testutil"
	"github.com/nelsw/bytelyon-client/pkg/url"
)

func setup(t *testing.T, depth int, code int) *testutil.Pool {
	t.Helper()
	p := testutil.DB(t)
	testutil.Redis(t)
	testutil.Workdir(t)
	testutil.Script(t, "sync_sitemap", code)

	orig := maxDepth
	maxDepth = depth
	t.Cleanup(func() { maxDepth = orig })
	return p
}

// filter returns the calls whose SQL contains s.
func filter(calls []testutil.Call, s string) (out []testutil.Call) {
	for _, c := range calls {
		if strings.Contains(c.SQL, s) {
			out = append(out, c)
		}
	}
	return
}

// finished waits for the bot update that marks the crawl as done and returns every exec up to then.
func finished(t *testing.T, p *testutil.Pool) []testutil.Call {
	t.Helper()
	testutil.Eventually(t, "bot update", func() bool { return len(filter(p.Execs(), "UPDATE bots")) > 0 })
	calls := p.Execs()
	if bots := filter(calls, "UPDATE bots"); len(bots) != 1 || bots[0].Args["id"] != 2 {
		t.Errorf("bot updates = %+v, want one for bot 2", bots)
	}
	return calls
}

func TestFetchNoSitemap(t *testing.T) {
	p := testutil.DB(t)

	Fetch(1, 0, "example.com", true)

	if len(p.Execs()) != 0 {
		t.Error("nothing should be saved without a sitemap")
	}
}

func TestFetchCrawl(t *testing.T) {
	p := setup(t, 2, 0)

	testutil.Output(t, "https://example.com", `{
		"title": "Home",
		"meta": {"description": "d"},
		"links": ["https://example.com/a", "https://www.example.com/b", "https://other.com/x", "http://example.com/y"]
	}`)
	// home was already visited, and /c is at depth 0, so its links aren't followed
	testutil.Output(t, "https://example.com/a", `{"title": "A", "links": ["https://example.com", "https://example.com/c"]}`)
	testutil.Output(t, "https://example.com/c", `{"title": "C", "links": ["https://example.com/d"]}`)
	testutil.Output(t, "https://www.example.com/b", `{"title": "B"}`)

	Fetch(2, 8, "example.com", false)

	calls := finished(t, p)
	want := []string{"https://example.com", "https://example.com/a", "https://example.com/c", "https://www.example.com/b"}

	updates := filter(calls, "UPDATE sitemaps")
	if len(updates) != 1 || updates[0].Args["id"] != 8 {
		t.Fatalf("sitemap updates = %+v, want exactly one", updates)
	}
	if got := updates[0].Args["urls"].([]string); !slices.Equal(got, want) {
		t.Errorf("urls = %v, want %v", got, want)
	}

	var args []string
	for _, u := range want {
		args = append(args, "-m false -u "+u)
	}
	if got := testutil.Args(t, "sync_sitemap"); got != strings.Join(args, "\n") {
		t.Errorf("sync_sitemap args = %q", got)
	}

	// pages are upserted concurrently, so order them by url
	pages := filter(calls, "INSERT INTO pages")
	slices.SortFunc(pages, func(a, b testutil.Call) int {
		return strings.Compare(a.Args["url"].(string), b.Args["url"].(string))
	})

	var urls []string
	for _, c := range pages {
		if c.Args["pageable_id"] != 8 || c.Args["domain"] != "example.com" {
			t.Errorf("unexpected upsert: %+v", c)
		}
		urls = append(urls, c.Args["url"].(string))
	}
	if !slices.Equal(urls, want) {
		t.Fatalf("pages = %v, want %v", urls, want)
	}
	if home := pages[0].Args; home["title"] != "Home" || home["meta"].(map[string]any)["description"] != "d" {
		t.Errorf("home page args = %v", home)
	}
	if b := pages[3].Args; b["meta"] == nil || b["screenshot_key"] != filepath.Join("sitemap", "8", url.UUID("https://www.example.com/b")+".png") {
		t.Errorf("page b args = %v", b)
	}
}

func TestFetchFailures(t *testing.T) {
	for name, tt := range map[string]struct {
		code int
		out  string
	}{
		"script fails":   {1, ""},
		"invalid output": {0, "not json"},
	} {
		t.Run(name, func(t *testing.T) {
			p := setup(t, 0, tt.code)
			testutil.Output(t, "https://example.com", tt.out)

			Fetch(2, 8, "example.com", false)

			// the crawl still finishes, saving an empty sitemap
			calls := finished(t, p)
			if updates := filter(calls, "UPDATE sitemaps"); len(updates) != 1 || updates[0].Args["urls"] != nil {
				t.Errorf("only a bare sitemap update expected, got %+v", calls)
			}
			if pages := filter(calls, "INSERT INTO pages"); len(pages) != 0 {
				t.Errorf("no pages expected, got %+v", pages)
			}
		})
	}
}

func TestRepoErrors(t *testing.T) {
	p := testutil.DB(t)
	p.ExecErr = errors.New("boom")

	UpsertPage(1, "d", "u", "t", "k", nil) // logs errors
	UpdateSitemap(1, []string{"u"})

	if calls := p.Execs(); len(calls) != 2 || calls[0].Args["meta"] == nil {
		t.Errorf("execs = %+v", calls)
	}
}
