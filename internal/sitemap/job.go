package sitemap

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/cache"
	"github.com/nelsw/bytelyon-client/pkg/fs"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/nelsw/bytelyon-client/pkg/url"
	"github.com/rs/zerolog/log"
)

type Job struct {
	id        int
	headless  bool
	domain    string
	frequency model.Frequency
	url       string
	depth     int
}

func (j *Job) String() string {
	return fmt.Sprintf("sitemap::%d::%s", j.id, j.url)
}

func (j *Job) Validate() bool {
	if url.Domain(j.url) != j.domain {
		return false
	}
	if cache.GetTime(j.String()).Add(j.frequency.Duration()).Before(time.Now()) {
		return false
	}
	return true
}

func (j *Job) Name() string {
	return "./scripts/sync_sitemap"
}

func (j *Job) Args() []string {
	return []string{
		"-m", strconv.FormatBool(j.headless),
		"-u", j.url,
	}
}

func (j *Job) Failure(err error) {
	log.Err(err).Stringer("job", j).Msg("failed to sync sitemap")
}

func (j *Job) Success(out []byte) {

	var page model.Page
	if err := json.Unmarshal(out, &page); err != nil {
		log.Err(err).Stringer("job", j).Msg("failed to unmarshal sitemap page")
		return
	}

	key := filepath.Join(
		string(bot.SitemapType),
		strconv.Itoa(j.id),
		uuid.NewSHA1(uuid.NameSpaceURL, []byte(j.url)).String(),
	) + ".png"
	_ = fs.Put(key, page.ScreenshotBytes())
	UpsertPage(j.id, j.domain, j.url, page.Title, key, page.Meta)
	cache.PutTime(j.String(), time.Now())

	nextDepth := j.depth - 1
	if nextDepth < 0 {
		return
	}

	for _, link := range page.Links {
		play.It(&Job{
			j.id,
			j.headless,
			j.domain,
			j.frequency,
			url.Clean(link),
			nextDepth,
		})
	}
}
