package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nelsw/bytelyon-client/pkg/db"
)

// Call is a statement executed against a Pool.
type Call struct {
	SQL  string
	Args pgx.StrictNamedArgs
}

// Pool is an in-memory db.Pool that records executed statements and serves canned rows.
type Pool struct {
	mu       sync.Mutex
	execs    []Call
	queries  []string
	ExecErr  error
	QueryErr error
	Rows     *Rows
	Closed   bool
}

// DB installs a fresh Pool as the db client for the duration of the test.
func DB(t testing.TB) *Pool {
	t.Helper()
	Isolate()
	p := &Pool{}
	db.Use(p)
	t.Cleanup(func() { db.Use(nil) })
	return p
}

// Execs returns the statements executed so far.
func (p *Pool) Execs() []Call {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]Call(nil), p.execs...)
}

// Queries returns the queries run so far.
func (p *Pool) Queries() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.queries...)
}

func (p *Pool) Query(_ context.Context, sql string, _ ...any) (pgx.Rows, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.queries = append(p.queries, sql)
	if p.QueryErr != nil {
		return nil, p.QueryErr
	} else if p.Rows == nil {
		return NewRows(nil), nil
	}
	return p.Rows, nil
}

func (p *Pool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	rows, err := p.Query(ctx, sql, args...)
	return row{rows, err}
}

func (p *Pool) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	c := Call{SQL: sql}
	if len(args) > 0 {
		c.Args, _ = args[0].(pgx.StrictNamedArgs)
	}
	p.execs = append(p.execs, c)
	return pgconn.NewCommandTag("OK"), p.ExecErr
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Closed = true
}

type row struct {
	rows pgx.Rows
	err  error
}

func (r row) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	defer r.rows.Close()
	if !r.rows.Next() {
		return pgx.ErrNoRows
	}
	return r.rows.Scan(dest...)
}

// Rows is a canned pgx.Rows result set.
type Rows struct {
	cols   []string
	values [][]any
	i      int
	Error  error
}

// NewRows returns a result set with the given column names and rows of values.
func NewRows(cols []string, values ...[]any) *Rows {
	return &Rows{cols: cols, values: values, i: -1}
}

func (r *Rows) Close()                        {}
func (r *Rows) Err() error                    { return r.Error }
func (r *Rows) CommandTag() pgconn.CommandTag { return pgconn.NewCommandTag("SELECT") }
func (r *Rows) RawValues() [][]byte           { return nil }
func (r *Rows) Conn() *pgx.Conn               { return nil }
func (r *Rows) TypeMap() *pgtype.Map          { return pgtype.NewMap() }

func (r *Rows) FieldDescriptions() []pgconn.FieldDescription {
	fds := make([]pgconn.FieldDescription, len(r.cols))
	for i, c := range r.cols {
		fds[i] = pgconn.FieldDescription{Name: c}
	}
	return fds
}

func (r *Rows) Next() bool {
	r.i++
	return r.i < len(r.values)
}

func (r *Rows) Values() ([]any, error) {
	return r.values[r.i], nil
}

// Scan assigns each value to its destination, honoring sql.Scanner and nil-able pointer destinations.
func (r *Rows) Scan(dest ...any) error {
	if r.i < 0 || r.i >= len(r.values) {
		return fmt.Errorf("scan: no current row")
	}
	if len(dest) == 1 {
		if rs, ok := dest[0].(pgx.RowScanner); ok {
			return rs.ScanRow(r)
		}
	}
	vals := r.values[r.i]
	if len(dest) != len(vals) {
		return fmt.Errorf("scan: %d destinations for %d values", len(dest), len(vals))
	}
	for i, d := range dest {
		v := vals[i]
		if s, ok := d.(sql.Scanner); ok {
			if err := s.Scan(v); err != nil {
				return err
			}
			continue
		}
		dv := reflect.ValueOf(d).Elem()
		if v == nil {
			dv.SetZero()
			continue
		}
		rv := reflect.ValueOf(v)
		switch {
		case rv.Type().AssignableTo(dv.Type()):
			dv.Set(rv)
		case dv.Kind() == reflect.Pointer && rv.Type().AssignableTo(dv.Type().Elem()):
			p := reflect.New(dv.Type().Elem())
			p.Elem().Set(rv)
			dv.Set(p)
		default:
			return fmt.Errorf("scan: cannot assign %T to %s", v, dv.Type())
		}
	}
	return nil
}
