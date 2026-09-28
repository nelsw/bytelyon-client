package sitemap

import (
	"os"
	"strconv"

	"github.com/nelsw/bytelyon-client/pkg/model"
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
	frequency model.Frequency,
) {

	if sitemapID == 0 {
		log.Warn().Int("bot", botID).Msg("no sitemap for bot")
		return
	}

	go play.It(&Job{
		sitemapID,
		headless,
		domain,
		frequency,
		"https://" + domain,
		maxDepth,
	})
}
