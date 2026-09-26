package postgres

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nelsw/bytelyon-client/pkg/ssh"
)

var client *pgxpool.Pool

func init() {

	ctx := context.Background()

	cfg, err := pgxpool.ParseConfig(os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}
	cfg.ConnConfig.DialFunc = ssh.DialFunc()

	if client, err = pgxpool.NewWithConfig(ctx, cfg); err != nil {
		panic(err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err = client.Ping(pingCtx); err != nil {
		panic(err)
	}
}

func Close() {
	client.Close()
}

func QueryRow(ctx context.Context, sql string, args []any, dest ...any) error {
	return client.QueryRow(ctx, sql, args...).Scan(dest...)
}

func Query(sql string, args ...any) (pgx.Rows, error) {
	return client.Query(context.Background(), sql, args...)
}

// Exec uses strict named args so a missing or misspelled arg errors instead of silently writing NULL.
func Exec(sql string, args pgx.StrictNamedArgs) error {
	_, err := client.Exec(context.Background(), sql, args)
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
  AND bots.type = ANY (@types)
  AND (due.at IS NULL OR due.at <= NOW())
ORDER BY due.at NULLS FIRST, bots.id
`

func SearchBots(types []string) (pgx.Rows, error) {
	return client.Query(context.Background(), DueBots, pgx.NamedArgs{"types": types})
}

// Varchar trims s to at most n characters so it fits a varchar(n) column.
func Varchar(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}
