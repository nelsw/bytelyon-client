package search

import (
	"github.com/jackc/pgx/v5"
	"github.com/nelsw/bytelyon-client/pkg/db"
	"github.com/rs/zerolog/log"
)

func UpdateSearch(searchID int, imgKey, srcKey string, data map[string]any) {
	sql := `
UPDATE serps
    SET data = @data,
        screenshot_key = @screenshot_key,
        content_key = @content_key,
        updated_at = NOW()
WHERE id = @id
`
	d := pgx.StrictNamedArgs{
		"id":             searchID,
		"data":           data,
		"screenshot_key": imgKey,
		"content_key":    srcKey,
	}
	if err := db.Exec(sql, d); err != nil {
		log.Warn().Err(err).Msg("failed to save serp")
	} else {
		log.Info().Msg("serp saved")
	}
}
