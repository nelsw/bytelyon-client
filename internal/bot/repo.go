package bot

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/nelsw/bytelyon-client/pkg/db"
	"github.com/rs/zerolog/log"
)

const findSQL = `
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

func FindOne() (m Model) {
	rows, err := db.Query(findSQL + "LIMIT 1")
	if err != nil {
		log.Err(err).Msg("failed to query bots")
		return
	}
	defer rows.Close()

	if m, err = pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[Model]); errors.Is(err, pgx.ErrNoRows) {
		log.Debug().Msg("no bots found")
		return
	} else if err != nil {
		log.Err(err).Msg("failed to collect bot")
		return Model{} // discard the partially scanned row
	}
	return
}

func FindAll() (arr []Model) {
	rows, err := db.Query(findSQL)
	if err != nil {
		log.Err(err).Msg("failed to query bots")
		return
	}
	defer rows.Close()

	if arr, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[Model]); errors.Is(err, pgx.ErrNoRows) {
		log.Debug().Msg("no bots found")
		return nil
	} else if err != nil {
		log.Err(err).Msg("failed to collect bots")
		return nil
	}
	return arr
}

func Update(m Model) {
	sql := `
UPDATE bots
SET last_run_at = NOW()
WHERE id = @id;
`
	if err := db.Exec(sql, pgx.StrictNamedArgs{"id": m.ID}); err != nil {
		log.Warn().Err(err).Msg("failed to update bot")
	}
}

func UpdateFn(m Model) func() {
	return func() { Update(m) }
}
