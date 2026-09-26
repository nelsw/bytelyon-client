package search

import (
	"os"

	"github.com/goforj/godump"
	"github.com/jackc/pgx/v5"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/postgres"
	"github.com/rs/zerolog/log"
)

func Save(id int, data model.Data[string, any], imgKey, srcKey string) {
	sql := `
UPDATE serps
    SET data = @data,
        screenshot_key = @screenshot_key,
        content_key = @content_key,
        updated_at = NOW()
WHERE id = @id
`
	d := pgx.StrictNamedArgs{
		"id":             id,
		"data":           map[string]any(data),
		"screenshot_key": imgKey,
		"content_key":    srcKey,
	}
	if err := postgres.Exec(sql, d); err != nil {
		log.Warn().Err(err).Msg("failed to save serp")
		if os.Getenv("APP_MODE") == "test" {
			godump.DumpJSON(d)
		}
	}
}
