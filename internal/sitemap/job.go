package sitemap

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/cache"
	"github.com/nelsw/bytelyon-client/pkg/fs"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/nelsw/bytelyon-client/pkg/url"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Job struct {
	id       int
	headless bool
	domain   string
	url      string
	depth    int
	start    time.Time
}

func (j *Job) s3(ext string) string {
	return filepath.Join(
		string(bot.SitemapType),
		strconv.Itoa(j.id),
		uuid.NewSHA1(uuid.NameSpaceURL, []byte(j.url)).String(),
	) + "." + ext
}

func (j *Job) db() string {
	return fmt.Sprintf("sitemap:%d:%s", j.id, j.url)
}

func (j *Job) MarshalZerologObject(evt *zerolog.Event) {
	evt.Int("#", j.id).
		Str("t", string(bot.SitemapType)).
		Str("u", j.url)
}

func (j *Job) Validate() bool {
	if !strings.HasPrefix(j.url, "https://") {
		return false
	}
	if url.Domain(j.url) != j.domain {
		return false
	}
	if cache.GetPage(j.db()).After(j.start) {
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
	log.Err(err).EmbedObject(j).Send()
}

func (j *Job) Success(out []byte) {

	var page model.Page
	if err := json.Unmarshal(out, &page); err != nil {
		log.Err(err).EmbedObject(j).Msg("failed to unmarshal sitemap page")
		return
	}

	key := j.s3("png")
	_ = fs.Put(key, page.ScreenshotBytes())
	UpsertPage(j.id, j.domain, j.url, page.Title, key, page.Meta)
	cache.PutPage(j.db())

	log.Debug().EmbedObject(j).Msg(`✅`)

	if j.depth-1 < 0 {
		return
	}

	for _, link := range page.Links {
		play.Go(&Job{
			j.id,
			j.headless,
			j.domain,
			url.Clean(link),
			j.depth - 1,
			j.start,
		})
	}
}
