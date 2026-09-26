package news

import (
	"os"

	"github.com/goforj/godump"
	"github.com/jackc/pgx/v5"
	"github.com/nelsw/bytelyon-client/pkg/postgres"
	"github.com/rs/zerolog/log"
)

func Save(botID int, a *Article) {
	sql := `
INSERT INTO articles (bot_id, url, title, published_at, img_alt, img_url, source, description, body, created_at, updated_at, publisher, keywords)
VALUES (@bot_id, @url, @title, @published_at, @img_alt, @img_url, @source, @description, @body, NOW(), NOW(), @publisher, @keywords)
ON CONFLICT (url, bot_id)
DO UPDATE SET title = EXCLUDED.title,
              img_alt = EXCLUDED.img_alt,
              img_url = EXCLUDED.img_url,
              source = EXCLUDED.source,
              description = EXCLUDED.description,
              body = EXCLUDED.body,
              publisher = EXCLUDED.publisher,
              keywords = EXCLUDED.keywords,
              updated_at = NOW()
`
	d := pgx.StrictNamedArgs{
		"bot_id":       botID,
		"url":          a.URL,
		"title":        postgres.Varchar(a.Title, 255),
		"published_at": a.PublishedAt(),
		"img_alt":      postgres.Varchar(a.ImgAlt, 1024),
		"img_url":      a.ImgURL,
		"source":       postgres.Varchar(a.Source, 255),
		"description":  a.Desc,
		"body":         a.Body,
		"publisher":    postgres.Varchar(a.Publisher, 255),
		"keywords":     a.Keywords,
	}
	if err := postgres.Exec(sql, d); err != nil {
		log.Warn().Err(err).Msg("failed to save article")
		if os.Getenv("APP_MODE") == "test" {
			godump.DumpJSON(d)
		}
	}
}
