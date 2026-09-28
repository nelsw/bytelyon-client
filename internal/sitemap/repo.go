package sitemap

import (
	"github.com/jackc/pgx/v5"
	"github.com/nelsw/bytelyon-client/pkg/db"
	"github.com/rs/zerolog/log"
)

const pageableType = "App\\Models\\Sitemap"

func UpsertPage(pid int, domain, url, title, imgKey string, meta map[string]any) {
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
	d := map[string]any{
		"domain":         domain,
		"meta":           meta,
		"pageable_id":    pid,
		"pageable_type":  pageableType,
		"screenshot_key": imgKey,
		"url":            url,
		"title":          db.Varchar(title, 1024),
	}

	if err := db.Exec(sql, pgx.StrictNamedArgs(d)); err != nil {
		log.Err(err).Msgf("failed to save page: %s", d)
	} else {
		log.Trace().Msgf("saved page: %s", d)
	}
}

func UpdateSitemap(id int, urls []string) {
	d := pgx.StrictNamedArgs{"id": id}
	sql := ` UPDATE sitemaps SET updated_at = NOW() WHERE id = @id`

	if len(urls) > 0 {
		d["urls"] = urls
		sql = `
UPDATE sitemaps
SET urls = @urls,
updated_at = NOW()
WHERE id = @id
`
	}

	if err := db.Exec(sql, d); err != nil {
		log.Err(err).Msgf("failed to save sitemap: %d", id)
	} else {

	}
}
