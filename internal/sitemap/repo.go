package sitemap

import (
	"os"

	"github.com/goforj/godump"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/postgres"
	"github.com/rs/zerolog/log"
)

func SavePage(m Model, d model.Data) {
	sql := `
INSERT INTO pages (url, domain, title, screenshot_key, meta, pageable_type, pageable_id, created_at, updated_at)
VALUES (@url, @domain, @screenshot_key, @meta, @pageable_type, @pageable_id, NOW(), NOW())
ON CONFLICT (pageable_type, pageable_id, url)
DO UPDATE SET title = @title, 
              meta = @meta, 
              screenshot_key = @screenshot_key,
              updated_at = NOW()
`
	err := postgres.Exec(sql, model.Data{
		"url":            d.Get("url"),
		"domain":         m.Bot.Query,
		"title":          d.Get("title"),
		"screenshot_key": d.Get("screenshot_key"),
		"meta":           d.Get("meta"),
		"pageable_type":  `App\Models\Sitemap`,
		"pageable_id":    m.ID,
	})

	if err != nil {
		log.Warn().Err(err).Msg("failed to save page")
		if os.Getenv("APP_MODE") == "test" {
			godump.DumpJSON(m, d)
		}
	}
}

func Save(id int, urls []string) {
	sql := `
UPDATE sitemaps
SET urls = @urls,
updated_at = NOW()
WHERE id = @id
`
	d := model.Data{
		"id":   id,
		"urls": urls,
	}

	err := postgres.Exec(sql, d)

	if err != nil {
		log.Warn().Err(err).Msg("failed to save sitemap")
		if os.Getenv("APP_MODE") == "test" {
			godump.DumpJSON(d)
		}
	}
}
