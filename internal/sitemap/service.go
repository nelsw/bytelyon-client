package sitemap

import (
	"os"
	"strconv"
	"time"

	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/rs/zerolog/log"
)

var maxDepth int

func init() {
	maxDepth, _ = strconv.Atoi(os.Getenv("SITEMAP_DEPTH"))
}

func Fetch(
	botID int,
	sitemapID int,
	domain string,
	headless bool,
) {

	if sitemapID == 0 {
		log.Warn().Int("bot", botID).Msg("no sitemap for bot")
		return
	}

	remove(botID)

	play.Go(&Job{
		botID,
		sitemapID,
		headless,
		domain,
		"https://" + domain,
		maxDepth,
		time.Now().Add(time.Second * -1),
	})
}
