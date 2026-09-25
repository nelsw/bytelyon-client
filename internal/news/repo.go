package news

import (
	"os"

	"github.com/goforj/godump"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/postgres"
	"github.com/rs/zerolog/log"
)

func Save(d model.Data) {
	sql := `
INSERT INTO articles (bot_id, url, title, published_at, img_alt, img_url, source, description, body, created_at, updated_at, publisher, keywords)
VALUES (@bot_id, @url, @title, @published_at, @img_alt, @img_url, @source, @description, @body, NOW(), NOW(), @publisher, @keywords)
ON CONFLICT (bot_id, url)
DO UPDATE SET title = @title,
              img_alt = @img_alt,
              img_url = @img_url,
              source = @source,
              description = @description,
              body = @body,
              publisher = @publisher,
              keywords = @keywords,
              updated_at = NOW()
`
	if err := postgres.Exec(sql, d); err != nil {
		log.Warn().Err(err).Msg("failed to save article")
		if os.Getenv("APP_MODE") == "test" {
			godump.DumpJSON(d)
		}
	}
}
