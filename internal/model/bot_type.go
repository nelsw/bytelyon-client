package model

import (
	"fmt"
	"strings"
)

type BotType string

const (
	NewsBot    BotType = "news"
	SearchBot  BotType = "search"
	SitemapBot BotType = "sitemap"
)

func (t *BotType) Class() string {
	switch *t {
	case NewsBot:
		return "App\\Models\\Article"
	case SearchBot:
		return "App\\Models\\Serp"
	case SitemapBot:
		return "App\\Models\\Sitemap"
	}
	return "unknown"
}

func (t *BotType) UnmarshalJSON(payload []byte) error {
	if text := strings.ReplaceAll(string(payload), `"`, ""); botTypeRegex.MatchString(text) {
		*t = BotType(text)
		return nil
	}
	return fmt.Errorf("unknown bot type: %s", payload)
}
