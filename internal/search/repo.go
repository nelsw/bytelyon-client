package search

import (
	"os"

	"github.com/goforj/godump"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/postgres"
	"github.com/rs/zerolog/log"
)

func Save(d model.Data) {
	sql := `
UPDATE serps
    SET data = @data,
        screenshot_key = @screenshot_key,
        content_key = @content_key
WHERE id = @id
`
	if err := postgres.Exec(sql, d); err != nil {
		log.Warn().Err(err).Msg("failed to save article")
		if os.Getenv("APP_MODE") == "test" {
			godump.DumpJSON(d)
		}
	}
}
