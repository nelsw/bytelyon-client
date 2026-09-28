package db

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nelsw/bytelyon-client/pkg/ssh"
)

// Pool is the subset of *pgxpool.Pool used by this package.
type Pool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Close()
}

var (
	client Pool
	mu     sync.Mutex
)

func connStr() string {
	if s := os.Getenv("POSTGRES_URL"); s != "" {
		return s
	}
	u, p := "root", "secret"
	if os.Getenv("APP_ENV") == "prod" {
		u, p = os.Getenv("DB_USER"), os.Getenv("DB_PASS")
	}
	return fmt.Sprintf("postgres://%s:%s@127.0.0.1:5432/forge?sslmode=disable", u, p)
}

// connect lazily opens the pool on first use; like before, a misconfigured database is fatal.
func connect() Pool {
	mu.Lock()
	defer mu.Unlock()

	if client != nil {
		return client
	}

	ctx := context.Background()

	cfg, err := pgxpool.ParseConfig(connStr())
	if err != nil {
		panic(err)
	}

	if os.Getenv("APP_ENV") == "prod" {
		cfg.ConnConfig.DialFunc = ssh.DialFunc()
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		panic(err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err = pool.Ping(pingCtx); err != nil {
		pool.Close()
		panic(err)
	}

	client = pool
	return client
}

// Use replaces the pool, e.g. with a mock in tests.
func Use(p Pool) {
	mu.Lock()
	defer mu.Unlock()
	client = p
}

func Close() {
	mu.Lock()
	defer mu.Unlock()

	if client != nil {
		client.Close()
		client = nil
	}
}

func QueryRow(ctx context.Context, sql string, args []any, dest ...any) error {
	return connect().QueryRow(ctx, sql, args...).Scan(dest...)
}

func Query(sql string, args ...any) (pgx.Rows, error) {
	return connect().Query(context.Background(), sql, args...)
}

// Exec uses strictly named args so missing or misspelled arg errors instead of silently writing NULL.
func Exec(sql string, args pgx.StrictNamedArgs) error {
	_, err := connect().Exec(context.Background(), sql, args)
	return err
}

// DueBots selects enabled bots that are due to run, never-run and most overdue first.
// child_id is the id of the bot's serps/sitemaps row (0 for news bots, or when no row exists).
const DueBots = `
SELECT bots.id,
       bots.type,
       bots.blacklist,
       bots.headless,
       bots.query,
       bots.last_run_at,
       COALESCE(CASE bots.type
                    WHEN 'search' THEN (SELECT serps.id
                                        FROM serps
                                        WHERE serps.bot_id = bots.id
                                          AND serps.deleted_at IS NULL
                                        ORDER BY serps.id DESC
                                        LIMIT 1)
                    WHEN 'sitemap' THEN (SELECT sitemaps.id
                                         FROM sitemaps
                                         WHERE sitemaps.bot_id = bots.id
                                           AND sitemaps.domain = bots.query
                                           AND sitemaps.deleted_at IS NULL
                                         LIMIT 1)
                    END, 0) AS child_id
FROM bots
         CROSS JOIN LATERAL (SELECT bots.last_run_at + CASE bots.frequency
                                                           WHEN 'hourly' THEN INTERVAL '1 hour'
                                                           WHEN 'daily' THEN INTERVAL '1 day'
                                                           WHEN 'weekly' THEN INTERVAL '1 week'
                                                           WHEN 'monthly' THEN INTERVAL '1 month'
                                                           ELSE INTERVAL '0'
                                                           END AS at) AS due
WHERE bots.enabled IS TRUE
  AND (due.at IS NULL OR due.at <= NOW())
ORDER BY due.at NULLS FIRST, bots.id
`

// Varchar trims s to at most n characters so it fits a varchar(n) column.
func Varchar(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}
