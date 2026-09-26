package search

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/nelsw/bytelyon-client/pkg/s3"
	"github.com/rs/zerolog/log"
)

func Build(b *bot.Model) {

	err := play.Search(b.ID, b.Query, b.Headless)
	if err != nil {
		log.Warn().Err(err).Msg("failed to scrape page")
		return
	}

	p := filepath.Join(".storage", "search", strconv.Itoa(b.ID), b.Query)

	from := p + ".html"
	srcKey := strings.ReplaceAll(from, ".storage/", "")
	_ = s3.Move(from, srcKey)
	//_ = os.Remove(from)

	from = p + ".png"
	imgKey := strings.ReplaceAll(from, ".storage/", "")
	_ = s3.Move(from, imgKey)
	//_ = os.Remove(from)

	var bytes []byte
	if bytes, err = os.ReadFile(p + ".json"); err != nil {
		log.Warn().Err(err).Msg("failed to read serp")
		return
	}

	var d model.Data[string, any]
	if err = json.Unmarshal(bytes, &d); err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal serp")
		return
	}

	Save(b.ChildID, d, imgKey, srcKey)
}
