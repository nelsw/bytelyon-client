package sitemap

import (
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/nelsw/bytelyon-client/pkg/cache"
	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/rs/zerolog/log"
)

var maxDepth int

func init() {
	maxDepth, _ = strconv.Atoi(os.Getenv("SITEMAP_DEPTH"))
}

func Fetch(
	sitemapID int,
	botID int,
	domain string,
	headless bool,
) {

	if sitemapID == 0 {
		log.Warn().Int("bot", botID).Msg("no sitemap for bot")
		return
	}

	play.Go(&Job{
		sitemapID,
		headless,
		domain,
		"https://" + domain,
		maxDepth,
		time.Now().Add(time.Second * -1),
	})

	prefix := "sitemap:" + strconv.Itoa(sitemapID) + ":"
	keys := func() []string { return cache.GetPageKeys(prefix + "*") }

	var lastCount int
	var wait func()
	wait = func() {

		newKeys := keys()
		thisCount := len(newKeys)

		log.Trace().
			Int("sitemap", sitemapID).
			Int("lastCount", lastCount).
			Int("thisCount", thisCount).
			Send()

		if lastCount == thisCount {
			var arr []string
			for _, key := range newKeys {
				arr = append(arr, strings.ReplaceAll(key, prefix, ""))
			}
			slices.Sort(arr)
			UpdateSitemap(sitemapID, arr...)
			return
		}

		lastCount = thisCount
		time.AfterFunc(time.Minute, wait)
	}
	time.Sleep(time.Minute)
	wait()
}
