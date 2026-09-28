package sitemap

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/cache"
)

func jobCountKey(botID int) string {
	return fmt.Sprintf("%s:%d", bot.SitemapType, botID)
}

func pageTimeKey(botID int, url string) string {
	return fmt.Sprintf("%s:%s", jobCountKey(botID), url)
}

func remove(botID int) {
	_ = cache.Del(jobCountKey(botID))
}
func incr(botID int) {
	_ = cache.Incr(jobCountKey(botID))
}
func decr(botID, id int) {
	jck := jobCountKey(botID)
	if i, _ := cache.Decr(jck); i > 0 {
		return
	}

	prefix := jck + ":"
	var urls []string
	for _, k := range cache.Keys(jck) {
		u, _ := strings.CutPrefix(k, prefix)
		urls = append(urls, u)
	}
	slices.Sort(urls)
	UpdateSitemap(id, urls)
	bot.Update(botID)
}

func lastVisit(botID int, url string) time.Time {
	return cache.GetTime(pageTimeKey(botID, url))
}

func saveVisit(botID int, id int, url string) {
	cache.SetTime(pageTimeKey(botID, url), time.Now())
	decr(botID, id)
}
