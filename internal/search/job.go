package search

import (
	"encoding/json"
	"path/filepath"
	"strconv"

	"github.com/nelsw/bytelyon-client/pkg/fs"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/rs/zerolog/log"
)

type Job struct {
	id       int
	headless bool
	query    string
}

func (j *Job) Name() string {
	return "./scripts/sync_search"
}

func (j *Job) Args() []string {
	return []string{
		"-q", j.query,
		"-m", strconv.FormatBool(j.headless),
	}
}

func (j *Job) Success(bytes []byte) {

	var p model.Page
	if err := json.Unmarshal(bytes, &p); err != nil {
		log.Err(err).Msg("failed to unmarshal page")
		return
	}

	key := filepath.Join("search", strconv.Itoa(j.id), j.query)
	_ = fs.Put(key+".png", p.ScreenshotBytes())
	_ = fs.Put(key+".html", p.ContentBytes())

	UpdateSearch(j.id, key+".png", key+".html", p.Data)
}

func (j *Job) Failure(err error) {
	log.Err(err).Msg("failed to scrape page")
}

func (j *Job) Validate() bool {
	return true
}
