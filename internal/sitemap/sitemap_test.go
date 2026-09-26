package sitemap

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon-client/internal/testutil"
)

// page writes the files the pages script would produce for url under sitemap id.
func page(t *testing.T, id, url, json string) {
	t.Helper()
	name := uuid.NewSHA1(uuid.NameSpaceURL, []byte(url)).String()
	testutil.Files(t, filepath.Join(".storage", "sitemap", id, name), json)
}

func withDepth(t *testing.T, d int) {
	orig := maxDepth
	maxDepth = d
	t.Cleanup(func() { maxDepth = orig })
}

func TestFetchNoSitemap(t *testing.T) {
	p := testutil.DB(t)

	Fetch(0, 1, "example.com", true)

	if len(p.Execs()) != 0 {
		t.Error("nothing should be saved without a sitemap")
	}
}

func TestFetchCrawl(t *testing.T) {
	p := testutil.DB(t)
	testutil.Workdir(t)
	testutil.Script(t, "pages", 0)
	withDepth(t, 1)

	page(t, "8", "https://example.com", `{
		"title": "Home",
		"meta": {"description": "d"},
		"links": ["https://example.com/a", "https://www.example.com/b", "https://other.com/x", 42, "https://example.com/a"]
	}`)
	page(t, "8", "https://example.com/a", `{"title": "A", "links": ["https://example.com", "https://example.com/c"]}`)
	page(t, "8", "https://www.example.com/b", `{"title": "B"}`)

	Fetch(8, 2, "example.com", false)

	if got, want := testutil.Args(t, "pages"), "-t sitemap -i 8 -m false -u https://example.com/a https://www.example.com/b"; got != want {
		t.Errorf("last pages args = %q, want %q", got, want)
	}

	calls := p.Execs()
	if len(calls) != 4 {
		t.Fatalf("expected 3 page upserts and 1 sitemap update, got %+v", calls)
	}

	var pages []string
	for _, c := range calls[:3] {
		if !strings.Contains(c.SQL, "INSERT INTO pages") || c.Args["pageable_id"] != 8 || c.Args["domain"] != "example.com" {
			t.Errorf("unexpected upsert: %+v", c)
		}
		pages = append(pages, c.Args["url"].(string))
	}
	if want := []string{"https://example.com", "https://example.com/a", "https://www.example.com/b"}; !slices.Equal(pages, want) {
		t.Errorf("pages = %v, want %v", pages, want)
	}
	if home := calls[0].Args; home["title"] != "Home" || home["meta"].(map[string]any)["description"] != "d" {
		t.Errorf("home page args = %v", home)
	}
	if b := calls[2].Args; b["meta"] == nil || b["screenshot_key"] != filepath.Join("sitemap", "8", uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://www.example.com/b")).String()+".png") {
		t.Errorf("page b args = %v", b)
	}

	update := calls[3]
	if !strings.Contains(update.SQL, "UPDATE sitemaps") || update.Args["id"] != 8 {
		t.Errorf("unexpected update: %+v", update)
	}
	want := []string{"https://example.com", "https://example.com/a", "https://example.com/c", "https://www.example.com/b"}
	if got := update.Args["urls"].([]string); !slices.Equal(got, want) {
		t.Errorf("urls = %v, want %v", got, want)
	}
}

func TestFetchNegativeDepth(t *testing.T) {
	p := testutil.DB(t)
	withDepth(t, -1)

	Fetch(8, 2, "example.com", false)

	if calls := p.Execs(); len(calls) != 1 || len(calls[0].Args["urls"].([]string)) != 0 {
		t.Errorf("only an empty sitemap update expected, got %+v", calls)
	}
}

func TestFetchScriptFailure(t *testing.T) {
	p := testutil.DB(t)
	testutil.Workdir(t)
	testutil.Script(t, "pages", 1)
	withDepth(t, 0)

	Fetch(8, 2, "example.com", false)

	if calls := p.Execs(); len(calls) != 1 || !strings.Contains(calls[0].SQL, "UPDATE sitemaps") {
		t.Errorf("only the sitemap update expected, got %+v", calls)
	}
}

func TestRepoErrors(t *testing.T) {
	p := testutil.DB(t)
	p.ExecErr = errors.New("boom")

	UpsertPage(1, "d", "u", "t", "k", nil) // logs errors
	UpdateSitemap(1, nil)

	if calls := p.Execs(); len(calls) != 2 || calls[0].Args["meta"] == nil {
		t.Errorf("execs = %+v", calls)
	}
}
