package db_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/nelsw/bytelyon-client/internal/testutil"
	"github.com/nelsw/bytelyon-client/pkg/db"
)

func TestExec(t *testing.T) {
	p := testutil.DB(t)

	args := pgx.StrictNamedArgs{"id": 1}
	if err := db.Exec("UPDATE x", args); err != nil {
		t.Fatal(err)
	}
	if calls := p.Execs(); len(calls) != 1 || calls[0].SQL != "UPDATE x" || calls[0].Args["id"] != 1 {
		t.Errorf("unexpected execs: %+v", calls)
	}

	p.ExecErr = errors.New("boom")
	if err := db.Exec("UPDATE x", args); !errors.Is(err, p.ExecErr) {
		t.Errorf("Exec() error = %v", err)
	}
}

func TestQuery(t *testing.T) {
	p := testutil.DB(t)
	p.Rows = testutil.NewRows([]string{"n"}, []any{7})

	rows, err := db.Query("SELECT n")
	if err != nil {
		t.Fatal(err)
	}
	n, err := pgx.CollectOneRow(rows, pgx.RowTo[int])
	if err != nil || n != 7 {
		t.Errorf("Query() row = %d, %v", n, err)
	}
}

func TestQueryRow(t *testing.T) {
	p := testutil.DB(t)
	p.Rows = testutil.NewRows([]string{"s"}, []any{"v"})

	var s string
	if err := db.QueryRow(context.Background(), "SELECT s", nil, &s); err != nil || s != "v" {
		t.Errorf("QueryRow() = %q, %v", s, err)
	}
}

func TestClose(t *testing.T) {
	p := testutil.DB(t)
	db.Close()
	if !p.Closed {
		t.Error("Close should close the pool")
	}
	db.Close() // no-op once closed
}

func TestConnectPanics(t *testing.T) {
	testutil.Isolate()
	for name, dsn := range map[string]string{
		"invalid url":     "postgres://%zz",
		"ping failure":    os.Getenv("POSTGRES_URL"),
		"bad pool config": "postgres://127.0.0.1:1/x?pool_max_conns=0",
	} {
		t.Run(name, func(t *testing.T) {
			db.Use(nil)
			t.Setenv("POSTGRES_URL", dsn)
			defer func() {
				if recover() == nil {
					t.Error("expected panic")
				}
			}()
			_, _ = db.Query("SELECT 1")
		})
	}
}

func TestConnStr(t *testing.T) {
	t.Setenv("POSTGRES_URL", "")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASS", "p")

	t.Setenv("APP_ENV", "local")
	if got, want := db.ConnStr(), "postgres://root:secret@127.0.0.1:5432/forge?sslmode=disable"; got != want {
		t.Errorf("local connStr() = %q, want %q", got, want)
	}
	t.Setenv("APP_ENV", "prod")
	if got, want := db.ConnStr(), "postgres://u:p@127.0.0.1:5432/forge?sslmode=disable"; got != want {
		t.Errorf("prod connStr() = %q, want %q", got, want)
	}
	t.Setenv("POSTGRES_URL", "postgres://override")
	if got := db.ConnStr(); got != "postgres://override" {
		t.Errorf("POSTGRES_URL should take precedence, got %q", got)
	}
}

func TestVarchar(t *testing.T) {
	for _, tt := range []struct {
		in   string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello", 3, "hel"},
		{"héllo", 2, "hé"},
	} {
		if got := db.Varchar(tt.in, tt.n); got != tt.want {
			t.Errorf("Varchar(%q, %d) = %q, want %q", tt.in, tt.n, got, tt.want)
		}
	}
}
