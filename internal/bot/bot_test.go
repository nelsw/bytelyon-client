package bot

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nelsw/bytelyon-client/internal/testutil"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/rs/zerolog"
)

var cols = []string{"id", "type", "blacklist", "headless", "query", "last_run_at", "child_id"}

func TestLastRun(t *testing.T) {
	if got := (&Model{}).LastRun(); !got.IsZero() {
		t.Errorf("LastRun() = %v, want zero", got)
	}
	now := time.Now()
	if got := (&Model{LastRunAt: &now}).LastRun(); !got.Equal(now) {
		t.Errorf("LastRun() = %v, want %v", got, now)
	}
}

func TestMarshalZerologObject(t *testing.T) {
	var buf bytes.Buffer
	l := zerolog.New(&buf)

	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	l.Log().EmbedObject(&Model{ID: 1, Query: "q", Type: NewsType, LastRunAt: &at}).Send()
	l.Log().EmbedObject(&Model{ID: 2}).Send()

	out := buf.String()
	for _, want := range []string{`"#":1`, `"q":"q"`, `"t":"news"`, `"@":"2026-01-02 03:04:05"`, `"#":2`, `"@":"Never"`} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %s: %s", want, out)
		}
	}
}

func TestTypeScan(t *testing.T) {
	for _, v := range []any{"news", "search", "sitemap"} {
		var typ Type
		if err := typ.Scan(v); err != nil {
			t.Errorf("Scan(%v) error = %v", v, err)
		}
	}
	var typ Type
	if err := typ.Scan("other"); err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestBlacklist(t *testing.T) {
	var b Blacklist
	if err := b.Scan(` Foo \nbar\n\nnull`); err != nil {
		t.Fatal(err)
	}
	s := model.Set[string](b)
	if got := s.Keys(); !slices.Equal(got, []string{"Foo", "bar"}) {
		t.Errorf("keys = %v", got)
	}
	if b.OK([]string{"bar", "baz"}) {
		t.Error("OK should reject blacklisted words")
	}
	if !b.OK([]string{"baz"}) {
		t.Error("OK should accept other words")
	}

	if err := b.Scan(nil); err != nil || !b.OK([]string{"bar"}) {
		t.Errorf("nil blacklist should allow everything, err = %v", err)
	}
}

func TestFindOne(t *testing.T) {
	p := testutil.DB(t)
	at := time.Now()
	p.Rows = testutil.NewRows(cols, []any{1, "search", nil, true, "golang", at, 9})

	m := FindOne()
	if m.ID != 1 || m.Type != SearchType || !m.Headless || m.Query != "golang" || m.ChildID != 9 || !m.LastRun().Equal(at) {
		t.Errorf("FindOne() = %+v", m)
	}
	if q := p.Queries(); len(q) != 1 || !strings.HasSuffix(q[0], "LIMIT 1") {
		t.Errorf("queries = %q", q)
	}
}

func TestFindOneFailures(t *testing.T) {
	tests := map[string]func(*testutil.Pool){
		"no rows":     func(p *testutil.Pool) { p.Rows = testutil.NewRows(cols) },
		"query error": func(p *testutil.Pool) { p.QueryErr = errors.New("boom") },
		"scan error":  func(p *testutil.Pool) { p.Rows = testutil.NewRows(cols, []any{1, "bogus", nil, true, "q", nil, 0}) },
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			setup(testutil.DB(t))
			if m := FindOne(); m.ID != 0 {
				t.Errorf("FindOne() = %+v, want zero", m)
			}
		})
	}
}

func TestFindAll(t *testing.T) {
	p := testutil.DB(t)
	p.Rows = testutil.NewRows(cols,
		[]any{1, "news", `a\nb`, false, "x", nil, 0},
		[]any{2, "sitemap", nil, true, "y.com", nil, 4},
	)

	arr := FindAll()
	if len(arr) != 2 || arr[0].ID != 1 || arr[1].Type != SitemapType || arr[1].ChildID != 4 {
		t.Errorf("FindAll() = %+v", arr)
	}
	if arr[0].Blacklist.OK([]string{"a"}) {
		t.Error("blacklist was not scanned")
	}
}

func TestFindAllFailures(t *testing.T) {
	tests := map[string]func(*testutil.Pool){
		"no rows":     func(p *testutil.Pool) { p.Rows = testutil.NewRows(cols) },
		"query error": func(p *testutil.Pool) { p.QueryErr = errors.New("boom") },
		"rows error":  func(p *testutil.Pool) { p.Rows = testutil.NewRows(cols); p.Rows.Error = errors.New("boom") },
	}
	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			setup(testutil.DB(t))
			if arr := FindAll(); len(arr) != 0 {
				t.Errorf("FindAll() = %+v, want none", arr)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	p := testutil.DB(t)

	UpdateFn(Model{ID: 5})()

	calls := p.Execs()
	if len(calls) != 1 || !strings.Contains(calls[0].SQL, "UPDATE bots") || calls[0].Args["id"] != 5 {
		t.Errorf("execs = %+v", calls)
	}

	p.ExecErr = errors.New("boom")
	Update(Model{ID: 5}) // logs a warning
	if len(p.Execs()) != 2 {
		t.Error("expected a second exec")
	}
}
