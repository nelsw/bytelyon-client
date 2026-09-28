package sitemap

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/emo"
	"github.com/nelsw/bytelyon-client/pkg/fs"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/nelsw/bytelyon-client/pkg/url"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Job struct {
	botID    int
	id       int
	headless bool
	domain   string
	url      string
	depth    int
	start    time.Time
}

func (j *Job) MarshalZerologObject(evt *zerolog.Event) {
	evt.Int("#", j.id).
		Str("t", string(bot.SitemapType)).
		Str("u", j.url)
}

func (j *Job) Validate() bool {

	ok := url.Secure(j.url) &&
		url.Domain(j.url) == j.domain &&
		lastVisit(j.botID, j.url).Before(j.start)

	if ok {
		incr(j.botID)
	}

	return ok
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
	decr(j.botID, j.id)
}

func (j *Job) Success(out []byte) {

	var page model.Page
	if err := json.Unmarshal(out, &page); err != nil {
		j.Failure(err)
		return
	}
	// deferred until the links are queued (and counted), so the count can't hit zero while the crawl continues
	defer decr(j.botID, j.id)

	key := fmt.Sprintf("%s/%d/%s.png", bot.SitemapType, j.id, url.UUID(j.url))
	_ = fs.Put(key, page.ScreenshotBytes())
	UpsertPage(j.id, j.domain, j.url, page.Title, key, page.Meta)

	log.Info().EmbedObject(j).Msg(emo.Truthy)
	saveVisit(j.botID, j.url)

	if j.depth-1 < 0 {
		return
	}

	for _, link := range page.Links {
		play.Go(&Job{
			j.botID,
			j.id,
			j.headless,
			j.domain,
			url.Clean(link),
			j.depth - 1,
			j.start,
		})
	}
}
