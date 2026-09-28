package sitemap

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon-client/internal/testutil"
)

func setup(t *testing.T, depth int, code int) *testutil.Pool {
	t.Helper()
	p := testutil.DB(t)
	testutil.Redis(t)
	testutil.Workdir(t)
	testutil.Script(t, "sync_sitemap", code)

	origDepth, origSettle := maxDepth, settle
	maxDepth, settle = depth, time.Second
	t.Cleanup(func() { maxDepth, settle = origDepth, origSettle })
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

// urlsUpdate waits for, and returns, the update that saves the sitemap's crawled urls.
func urlsUpdate(t *testing.T, p *testutil.Pool) testutil.Call {
	t.Helper()
	var update []testutil.Call
	testutil.Eventually(t, "sitemap urls update", func() bool {
		update = filter(p.Execs(), "SET urls")
		return len(update) > 0
	})
	return update[0]
}

func TestFetchNoSitemap(t *testing.T) {
	p := testutil.DB(t)

	Fetch(0, 1, "example.com", true)

	if len(p.Execs()) != 0 {
		t.Error("nothing should be saved without a sitemap")
	}
}

func TestFetchCrawl(t *testing.T) {
	p := setup(t, 1, 0)

	testutil.Output(t, "https://example.com", `{
		"title": "Home",
		"meta": {"description": "d"},
		"links": ["https://example.com/a", "https://www.example.com/b", "https://other.com/x", "http://example.com/y"]
	}`)
	testutil.Output(t, "https://example.com/a", `{"title": "A", "links": ["https://example.com", "https://example.com/c"]}`)
	testutil.Output(t, "https://www.example.com/b", `{"title": "B"}`)

	Fetch(8, 2, "example.com", false)

	update := urlsUpdate(t, p)
	if update.Args["id"] != 8 {
		t.Errorf("unexpected update: %+v", update)
	}
	// pages at depth 0 are saved but their links are not followed
	want := []string{"https://example.com", "https://example.com/a", "https://www.example.com/b"}
	if got := update.Args["urls"].([]string); !slices.Equal(got, want) {
		t.Errorf("urls = %v, want %v", got, want)
	}
	if got, want := testutil.Args(t, "sync_sitemap"), "-m false -u https://example.com\n-m false -u https://example.com/a\n-m false -u https://www.example.com/b"; got != want {
		t.Errorf("sync_sitemap args = %q, want %q", got, want)
	}

	// pages are upserted concurrently, so order them by url
	calls := filter(p.Execs(), "INSERT INTO pages")
	slices.SortFunc(calls, func(a, b testutil.Call) int {
		return strings.Compare(a.Args["url"].(string), b.Args["url"].(string))
	})

	var pages []string
	for _, c := range calls {
		if c.Args["pageable_id"] != 8 || c.Args["domain"] != "example.com" {
			t.Errorf("unexpected upsert: %+v", c)
		}
		pages = append(pages, c.Args["url"].(string))
	}
	if !slices.Equal(pages, want) {
		t.Fatalf("pages = %v, want %v", pages, want)
	}
	if home := calls[0].Args; home["title"] != "Home" || home["meta"].(map[string]any)["description"] != "d" {
		t.Errorf("home page args = %v", home)
	}
	if b := calls[2].Args; b["meta"] == nil || b["screenshot_key"] != filepath.Join("sitemap", "8", uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://www.example.com/b")).String()+".png") {
		t.Errorf("page b args = %v", b)
	}
}

func TestFetchScriptFailure(t *testing.T) {
	p := setup(t, 0, 1)

	Fetch(8, 2, "example.com", false)

	testutil.Eventually(t, "script run", func() bool { _, err := os.Stat("sync_sitemap.args"); return err == nil })
	testutil.Eventually(t, "sitemap update", func() bool { return len(p.Execs()) > 0 })
	if calls := p.Execs(); len(calls) != 1 || !strings.Contains(calls[0].SQL, "UPDATE sitemaps") || calls[0].Args["urls"] != nil {
		t.Errorf("only a bare sitemap update expected, got %+v", calls)
	}
	if got := testutil.Args(t, "sync_sitemap"); got != "-m false -u https://example.com" {
		t.Errorf("sync_sitemap args = %q", got)
	}
}

func TestRepoErrors(t *testing.T) {
	p := testutil.DB(t)
	p.ExecErr = errors.New("boom")

	UpsertPage(1, "d", "u", "t", "k", nil) // logs errors
	UpdateSitemap(1, "")

	if calls := p.Execs(); len(calls) != 2 || calls[0].Args["meta"] == nil {
		t.Errorf("execs = %+v", calls)
	}
}
