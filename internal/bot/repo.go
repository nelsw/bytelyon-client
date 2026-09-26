package bot

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/nelsw/bytelyon-client/pkg/postgres"
	"github.com/rs/zerolog/log"
)

func FindOne() *Model {
	rows, err := postgres.Query(postgres.DueBots+"LIMIT 1", pgx.NamedArgs{
		"types": []string{
			string(NewsType),
			string(SearchType),
			string(SitemapType),
		},
	})
	if err != nil {
		log.Err(err).Msg("failed to query bots")
		return nil
	}
	defer rows.Close()

	var m Model
	if m, err = pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[Model]); errors.Is(err, pgx.ErrNoRows) {
		log.Debug().Msg("no bots found")
		return nil
	} else if err != nil {
		log.Err(err).Msg("failed to collect bot")
		return nil
	}
	return &m
}

func FindAll() (arr []Model) {
	rows, err := postgres.Query(postgres.DueBots, pgx.NamedArgs{
		"types": []string{
			string(NewsType),
			string(SearchType),
			string(SitemapType),
		},
	})
	if err != nil {
		log.Err(err).Msg("failed to query bots")
		return
	}
	defer rows.Close()

	if arr, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[Model]); err != nil {
		log.Err(err).Msg("failed to collect bots")
		return
	}
	return arr
}

func Update(id int) {
	sql := `
UPDATE bots
SET last_run_at = NOW()
WHERE id = @id;
`

	if err := postgres.Exec(sql, pgx.StrictNamedArgs{"id": id}); err != nil {
		log.Warn().Err(err).Msg("failed to update bot")
	}
}
