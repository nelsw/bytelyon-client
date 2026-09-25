package search

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/s3"
	"github.com/nelsw/bytelyon-client/pkg/scrape"
	"github.com/rs/zerolog/log"
)

func Build(b bot.Model) {

	key, err := scrape.Serp(b.Headless, b.Query, b.Type, b.ID)
	if err != nil {
		log.Warn().Err(err).Msg("failed to scrape page")
		return
	}

	from := key + ".png"
	imgKey := strings.ReplaceAll(from, ".storage", "bots")
	_ = s3.Move(from, imgKey)

	from = key + ".html"
	srcKey := strings.ReplaceAll(from, ".storage", "bots")
	_ = s3.Move(from, srcKey)

	var bytes []byte
	if bytes, err = os.ReadFile(key + ".json"); err != nil {
		log.Warn().Err(err).Msg("failed to read serp")
		return
	}

	var d model.Data
	if err = json.Unmarshal(bytes, &d); err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal serp")
		return
	}

	d.Put("id", b.ChildID)
	d.Put("screenshot_key", imgKey)
	d.Put("content_key", srcKey)

	Save(d)
}
