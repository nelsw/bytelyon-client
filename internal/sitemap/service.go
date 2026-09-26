package sitemap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/nelsw/bytelyon-client/pkg/s3"
	"github.com/nelsw/bytelyon-client/pkg/url"
	"github.com/rs/zerolog/log"
)

func Fetch(b *bot.Model) {
	if b.ChildID == 0 {
		log.Warn().Int("bot", b.ID).Msg("no sitemap for bot")
		return
	}
	urls := model.NewSyncSet[string]()
	if err := Scrape(6, b, urls, "https://"+b.Query); err != nil {
		log.Warn().Err(err).Msg("failed to scrape sitemap")
	}
	Save(b.ChildID, urls.Keys())
}

func Scrape(depth int, b *bot.Model, done *model.SyncSet[string], urls ...string) error {

	log.Debug().
		Int("depth", depth).
		Str("domain", b.Query).
		Int("urls", len(urls)).
		Msg("scraping sitemap")

	if depth < 0 || len(urls) == 0 {
		return nil
	}

	if err := play.Pages(b.Type, b.ID, b.Headless, urls); err != nil {
		log.Warn().Err(err).Msg("failed to scrape pages")
		return err
	}

	todo := model.NewSyncSet[string]()

	path := filepath.Join(".storage", "sitemap", strconv.Itoa(b.ID))

	ƒ := func(u string) func() {
		return func() {
			p := filepath.Join(path, uuid.NewSHA1(uuid.NameSpaceURL, []byte(u)).String())

			from := p + ".html"
			to := strings.ReplaceAll(from, ".storage/", "")
			_ = s3.Move(from, to)
			//_ = os.Remove(from)

			from = p + ".png"
			to = strings.ReplaceAll(from, ".storage/", "")
			_ = s3.Move(from, to)
			//_ = os.Remove(from)

			from = p + ".json"
			bytes, err := os.ReadFile(from)
			if err != nil {
				log.Warn().Err(err).Msg("failed to read page")
				return
			}
			//_ = os.Remove(from)

			var d model.Data[string, any]
			if err = json.Unmarshal(bytes, &d); err != nil {
				log.Warn().Err(err).Msg("failed to unmarshal page")
				return
			}

			title, _ := d.Get("title").(string)
			meta, _ := d.Get("meta").(map[string]any)

			SavePage(b.ChildID, b.Query, u, title, to, meta)
			done.Add(u)

			for _, link := range d.Get("links").([]any) {
				str := link.(string)
				if url.Domain(str) == b.Query && !done.Has(str) && !todo.Has(str) {
					todo.Add(str)
				}
			}
		}
	}

	for _, u := range urls {
		ƒ(u)
	}

	if depth > 0 {
		return Scrape(depth-1, b, done, todo.Keys()...)
	}

	done.AddAll(todo.Keys())
	return nil
}
