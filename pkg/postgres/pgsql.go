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

func QueryRow(ctx context.Context, sql string, args []any, dest any) error {
	return client.QueryRow(ctx, sql, args...).Scan(&dest)
}

func Query(sql string, args ...any) (pgx.Rows, error) {
	return client.Query(context.Background(), sql, args...)
}

func Exec(sql string, args map[string]any) error {
	_, err := client.Exec(context.Background(), sql, args)
	return err
}

func SearchBots(types []string) (pgx.Rows, error) {
	query := `
SELECT bss.id,
       bss.type,
       bss.blacklist,
       bss.headless,
       bss.query,
       CASE
           WHEN bss.type = 'search' THEN serp_id
           WHEN bss.type = 'sitemap' THEN sitemap_id
           ELSE 0
           END AS child_id
FROM (SELECT bots.id,
             bots.query,
             bots.headless,
             bots.type,
             bots.blacklist,
             bots.last_run_at,
             serps.bot_id    AS serp_id,
             sitemaps.bot_id AS sitemap_id,
             CASE
                 WHEN bots.frequency = 'monthly' THEN 60 * 24 * 7 * 52
                 WHEN bots.frequency = 'weekly' THEN 60 * 24 * 7
                 WHEN bots.frequency = 'daily' THEN 60 * 24
                 WHEN bots.frequency = 'hourly' THEN 60
                 ELSE 0
                 END         AS minutes
      FROM bots
               LEFT JOIN serps ON bots.id = serps.bot_id AND bots.type = 'search'
               LEFT JOIN sitemaps ON bots.id = sitemaps.bot_id AND bots.type = 'sitemap'
      WHERE bots.enabled IS TRUE) AS bss
WHERE bss.type = ANY($1) 
  AND (
    bss.last_run_at IS NULL OR
    bss.last_run_at + (bss.minutes * INTERVAL '1 minute') <= NOW()
    )
ORDER BY (NOW() - (bss.last_run_at + (bss.minutes * INTERVAL '1 minute'))) DESC;`

	return client.Query(context.Background(), query, types)
}
