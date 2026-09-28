package search

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/cache"
	"github.com/nelsw/bytelyon-client/pkg/fs"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/rs/zerolog/log"
)

type Job struct {
	botID    int
	searchID int
	headless bool
	query    string
}

func (j *Job) Name() string {
	return "./scripts/sync_search"
}

func (j *Job) Args() []string {
	return []string{
		"-m", strconv.FormatBool(j.headless),
		"-q", j.query,
	}
}

func (j *Job) Success(bytes []byte) {

	var p model.Page
	if err := json.Unmarshal(bytes, &p); err != nil {
		j.Failure(err)
		return
	}

	key := filepath.Join(string(bot.SearchType), strconv.Itoa(j.searchID), j.query)
	imgKey, srcKey := key+".png", key+".html"
	_ = fs.Put(imgKey, p.ScreenshotBytes())
	_ = fs.Put(srcKey, p.ContentBytes())

	UpdateSearch(j.searchID, imgKey, srcKey, p.Data)

	cache.Publish(j.botID, fmt.Sprintf("%s Search result ready!", j.query))
	bot.Update(j.botID)
}

func (j *Job) Failure(err error) {
	log.Err(err).Msg("failed to scrape page")
	bot.Update(j.botID)
}

func (j *Job) Validate() bool {
	return true
}
