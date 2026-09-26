package sitemap

import (
	"github.com/jackc/pgx/v5"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/postgres"
	"github.com/rs/zerolog/log"
)

const pageableType = "App\\Models\\Sitemap"

func SavePage(pid int, domain, url, title, imgKey string, meta map[string]any) {
	sql := `
INSERT INTO pages (
                   domain,
                   meta,
                   pageable_id,
                   pageable_type,
                   screenshot_key,
                   title,
                   url,
                   created_at,
                   updated_at
                   )
VALUES (
        @domain,
        @meta,
        @pageable_id,
        @pageable_type,
        @screenshot_key,
        @title,
        @url,
        NOW(),
        NOW()
        )
ON CONFLICT (
    pageable_type,
    pageable_id,
    url
    )
DO UPDATE SET title = excluded.title,
              meta = excluded.meta,
              screenshot_key = excluded.screenshot_key,
              updated_at = NOW()
`
	if meta == nil {
		meta = map[string]any{}
	}
	d := model.Data[string, any]{
		"domain":         domain,
		"meta":           meta,
		"pageable_id":    pid,
		"pageable_type":  pageableType,
		"screenshot_key": imgKey,
		"url":            url,
		"title":          postgres.Varchar(title, 1024),
	}

	if err := postgres.Exec(sql, pgx.StrictNamedArgs(d)); err != nil {
		log.Err(err).Msgf("failed to save page: %s", d)
	} else {
		log.Trace().Msgf("saved page: %s", d)
	}
}

func Save(id int, urls []string) {
	sql := `
UPDATE sitemaps
SET urls = @urls,
updated_at = NOW()
WHERE id = @id
`
	d := pgx.StrictNamedArgs{
		"id":   id,
		"urls": urls,
	}

	if err := postgres.Exec(sql, d); err != nil {
		log.Err(err).Msgf("failed to save sitemap: %d", id)
	}
}
