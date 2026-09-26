package app

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/internal/testutil"
	"github.com/nelsw/bytelyon-client/pkg/cache"
)

func setup(t *testing.T) *testutil.Pool {
	t.Helper()
	p := testutil.DB(t)
	testutil.Redis(t)
	testutil.Workdir(t)
	testutil.Transport(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "<rss><channel></channel></rss>")
	}))
	return p
}

func TestInit(t *testing.T) {
	setup(t)
	Init()
}

func TestHandleBots(t *testing.T) {
	p := setup(t)

	HandleBots([]bot.Model{
		{},
		{ID: 1, Type: bot.NewsType, Query: "golang"},
		{ID: 2, Type: bot.SearchType, ChildID: 5, Query: "golang"}, // no script installed, so the search fails fast
		{ID: 3, Type: bot.SitemapType, Query: "example.com"},       // no child sitemap
	})

	var ids []any
	for _, c := range p.Execs() {
		if !strings.Contains(c.SQL, "UPDATE bots") {
			t.Errorf("unexpected exec: %+v", c)
		}
		ids = append(ids, c.Args["id"])
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Errorf("updated bots = %v, want [1 2 3]", ids)
	}
	if n := working.Load(); n != 0 {
		t.Errorf("working = %d after handling, want 0", n)
	}
}

func TestClose(t *testing.T) {
	p := setup(t)
	cache.Put("k", "v") // opens the redis client

	working.Add(1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		working.Add(-1)
	}()

	Close()

	if !p.Closed {
		t.Error("Close should close the database")
	}
	if n := working.Load(); n != 0 {
		t.Error("Close should wait for in-flight work")
	}
}
