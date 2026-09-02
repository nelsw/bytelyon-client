package news

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/nelsw/bytelyon-client/config"
	"github.com/nelsw/bytelyon-client/pkg/http"
	"github.com/nelsw/bytelyon-client/pkg/store"
	"github.com/nelsw/bytelyon-client/pkg/uuid"
	"github.com/rs/zerolog/log"
)

func get(q string, t time.Time, m map[string]bool) (arr []*Article) {

	ƒ := func(u string) (rss RSS) {
		if out, err := http.Get(u, nil); err == nil {
			_ = xml.Unmarshal(out, &rss)
		}
		return
	}

	p := strings.ReplaceAll(q, " ", "+")
	b := ƒ(fmt.Sprintf("https://www.bing.com/news/search?format=rss&q=%s", p))
	g := ƒ(fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=en-US&gl=US&ceid=US:en", p))

	all := append(b.Channel.Articles, g.Channel.Articles...)
	for _, a := range all {
		if a.IsAfter(t) && !a.IsBlacklisted(m) {
			arr = append(arr, a)
		}
	}

	log.Info().
		Int("all", len(all)).
		Int("ok", len(arr)).
		Str("query", q).
		Msg("got news")

	return arr
}

func put(id int, a *Article) (err error) {
	name := uuid.FromURL(a.URL).String() + ".json"
	if err = store.Save(a, "news", id, name); err != nil {
		log.Err(err).EmbedObject(a).Msg("failed to save article")
	}
	if config.DryRun() {
		return
	}
	url := fmt.Sprintf("%s/api/bots/%d/articles", config.ApiHost(), id)
	if _, err = http.Put(url, a, config.AuthHeader()); err != nil {
		log.Err(err).EmbedObject(a).Msg("failed to save article")
	}
	return
}
