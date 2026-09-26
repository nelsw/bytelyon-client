package main

import (
	"strings"
	"testing"

	"github.com/nelsw/bytelyon-client/internal/testutil"
)

func TestRun(t *testing.T) {
	p := testutil.DB(t)
	testutil.Redis(t)
	testutil.Workdir(t)
	p.Rows = testutil.NewRows([]string{"id", "type", "blacklist", "headless", "query", "last_run_at", "child_id"},
		[]any{4, "sitemap", nil, true, "example.com", nil, 0})

	main()

	if q := p.Queries(); len(q) != 1 || !strings.HasSuffix(q[0], "LIMIT 1") {
		t.Errorf("queries = %q", q)
	}
	if calls := p.Execs(); len(calls) != 1 || calls[0].Args["id"] != 4 {
		t.Errorf("execs = %+v", calls)
	}
	if !p.Closed {
		t.Error("main should close the database on exit")
	}
}
