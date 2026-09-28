package testutil

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nelsw/bytelyon-client/pkg/cache"
	"github.com/nelsw/bytelyon-client/pkg/db"
)

func TestIsolate(t *testing.T) {
	s3 := Isolate()
	if s3 != Isolate() {
		t.Error("Isolate should reuse the fake S3 server")
	}
	if os.Getenv("SERVER_ADDR") != "" || os.Getenv("S3_BUCKET") != Bucket || os.Getenv("AWS_ENDPOINT_URL_S3") != s3.URL {
		t.Error("environment was not isolated")
	}

	put := func() int {
		req, _ := http.NewRequest(http.MethodPut, s3.URL+"/b/k", strings.NewReader("v"))
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		return res.StatusCode
	}
	if code := put(); code != http.StatusOK {
		t.Errorf("put = %d", code)
	}
	if b, ok := s3.Object("b/k"); !ok || string(b) != "v" {
		t.Errorf("object = %q, %v", b, ok)
	}
	s3.Fail(true)
	defer s3.Fail(false)
	if code := put(); code != http.StatusForbidden {
		t.Errorf("failing put = %d", code)
	}
}

func TestPool(t *testing.T) {
	p := DB(t)

	if _, err := db.Query("SELECT"); err != nil {
		t.Error(err)
	}

	type rec struct {
		N int        `db:"n"`
		S string     `db:"s"`
		P *time.Time `db:"p"`
		Z *int       `db:"z"`
	}
	now := time.Now()
	p.Rows = NewRows([]string{"n", "s", "p", "z"}, []any{1, "a", now, nil})
	rows, _ := db.Query("SELECT")
	if r, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[rec]); err != nil || r.N != 1 || r.S != "a" || !r.P.Equal(now) || r.Z != nil {
		t.Errorf("row = %+v, %v", r, err)
	}

	p.Rows = NewRows([]string{"n"}, []any{"not an int"})
	rows, _ = db.Query("SELECT")
	if _, err := pgx.CollectOneRow(rows, pgx.RowTo[int]); err == nil {
		t.Error("expected type mismatch error")
	}

	p.Rows = NewRows([]string{"n", "m"}, []any{1, 2})
	rows, _ = db.Query("SELECT")
	if vals, err := pgx.CollectOneRow(rows, pgx.RowToMap); err != nil || vals["m"] != 2 {
		t.Errorf("RowToMap = %v, %v", vals, err)
	}
	if rows.CommandTag().String() != "SELECT" || rows.RawValues() != nil || rows.Conn() != nil || rows.TypeMap() == nil {
		t.Error("unexpected rows metadata")
	}

	var n int
	if err := rows.Scan(&n); err == nil {
		t.Error("expected no current row error")
	}
	p.Rows = NewRows([]string{"n", "m"}, []any{1, 2})
	if rows, _ = db.Query("SELECT"); !rows.Next() {
		t.Fatal("expected a row")
	} else if err := rows.Scan(&n); err == nil {
		t.Error("expected destination count error")
	}

	p.Rows = NewRows([]string{"n"}, []any{1})
	if err := db.QueryRow(context.Background(), "SELECT", nil, &n); err != nil || n != 1 {
		t.Errorf("QueryRow = %d, %v", n, err)
	}
	p.Rows = NewRows([]string{"n"})
	if err := db.QueryRow(context.Background(), "SELECT", nil, &n); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("QueryRow error = %v", err)
	}
	p.QueryErr = errors.New("boom")
	if err := db.QueryRow(context.Background(), "SELECT", nil, &n); !errors.Is(err, p.QueryErr) {
		t.Errorf("QueryRow error = %v", err)
	}

	if err := db.Exec("UPDATE", pgx.StrictNamedArgs{"a": 1}); err != nil {
		t.Error(err)
	}
	if _, err := p.Exec(context.Background(), "DELETE"); err != nil {
		t.Error(err)
	}
	if calls := p.Execs(); len(calls) != 2 || calls[0].Args["a"] != 1 || calls[1].Args != nil {
		t.Errorf("execs = %+v", calls)
	}
	if q := p.Queries(); len(q) != 8 {
		t.Errorf("queries = %d, want 8", len(q))
	}
}

func TestRedis(t *testing.T) {
	m := Redis(t)
	cache.Put(13, "k", "v")
	if v, _ := m.DB(13).Get("k"); v != "v" {
		t.Errorf("value = %q", v)
	}
}

func TestFilesAndScripts(t *testing.T) {
	dir := Workdir(t)
	if wd, _ := os.Getwd(); !strings.HasSuffix(wd, dir[strings.LastIndex(dir, "/"):]) {
		t.Errorf("wd = %s, want %s", wd, dir)
	}

	Files(t, "a/b/c", `{}`)
	for _, ext := range []string{".html", ".png", ".json"} {
		if _, err := os.Stat("a/b/c" + ext); err != nil {
			t.Error(err)
		}
	}

	Script(t, "ok", 0)
	Script(t, "fail", 3)
	Output(t, "https://a.com/y", "out")
	if out, err := exec.Command("./scripts/ok", "-x", "https://a.com/y").Output(); err != nil || string(out) != "out" {
		t.Errorf("out = %q, %v", out, err)
	}
	if out, err := exec.Command("./scripts/ok", "-x", "z").Output(); err != nil || len(out) != 0 {
		t.Errorf("out = %q, %v", out, err)
	}
	if got := Args(t, "ok"); got != "-x https://a.com/y\n-x z" {
		t.Errorf("args = %q", got)
	}
	var exit *exec.ExitError
	if err := exec.Command("./scripts/fail").Run(); !errors.As(err, &exit) || exit.ExitCode() != 3 {
		t.Errorf("exit = %v", err)
	}
}

func TestTransport(t *testing.T) {
	Transport(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, r.Header.Get("X-Original-Host")+r.URL.Path)
	}))

	res, err := http.Get("https://example.com/path")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if b, _ := io.ReadAll(res.Body); string(b) != "example.com/path" {
		t.Errorf("body = %q", b)
	}
}
