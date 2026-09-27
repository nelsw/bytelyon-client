package news

import (
	"github.com/jackc/pgx/v5"
	"github.com/nelsw/bytelyon-client/pkg/db"
	"github.com/rs/zerolog/log"
)

func UpsertArticle(botID int, a *Article) {
	sql := `
INSERT INTO articles (bot_id, url, title, published_at, img_alt, img_url, source, description, body, created_at, updated_at, publisher, keywords)
VALUES (@bot_id, @url, @title, @published_at, @img_alt, @img_url, @source, @description, @body, NOW(), NOW(), @publisher, @keywords)
ON CONFLICT (url, bot_id)
DO UPDATE SET title = excluded.title,
              img_alt = excluded.img_alt,
              img_url = excluded.img_url,
              source = excluded.source,
              description = excluded.description,
              body = excluded.body,
              publisher = excluded.publisher,
              keywords = excluded.keywords,
              updated_at = NOW()
`
	d := pgx.StrictNamedArgs{
		"bot_id":       botID,
		"url":          a.URL,
		"title":        db.Varchar(a.Title, 255),
		"published_at": a.PublishedAt(),
		"img_alt":      db.Varchar(a.ImgAlt, 1024),
		"img_url":      a.ImgURL,
		"source":       db.Varchar(a.Source, 255),
		"description":  a.Desc,
		"body":         a.Body,
		"publisher":    db.Varchar(a.Publisher, 255),
		"keywords":     a.Keywords,
	}
	if err := db.Exec(sql, d); err != nil {
		log.Warn().Err(err).Msg("failed to save article")
	} else {
		log.Debug().EmbedObject(a).Msg("💾")
	}
}
