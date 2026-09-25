package bot

import (
	"github.com/jackc/pgx/v5"
	"github.com/nelsw/bytelyon-client/pkg/postgres"
	"github.com/rs/zerolog/log"
)

func Find() (Model, error) {
	sql := `
SELECT bss.id AS id,
       bss.type AS type,
       bss.blacklist AS blacklist,
       bss.headless AS headless,
       bss.query AS query,
       bss.last_run_at AS last_run_at,
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
WHERE bss.last_run_at IS NULL 
   OR bss.last_run_at + (bss.minutes * INTERVAL '1 minute') <= NOW()
ORDER BY (NOW() - (bss.last_run_at + (bss.minutes * INTERVAL '1 minute')));
`
	rows, err := postgres.Query(sql)
	if err != nil {
		return Model{}, err
	}
	defer rows.Close()
	return pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[Model])
}

func Update(id int) {
	sql := `
UPDATE bots
SET last_run_at = NOW()
WHERE id = @id;
`

	if err := postgres.Exec(sql, pgx.NamedArgs{"id": id}); err != nil {
		log.Warn().Err(err).Msg("failed to update bot")
	}
}
