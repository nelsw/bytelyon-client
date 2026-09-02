package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var botTypeRegex = regexp.MustCompile(`^(news|search|sitemap)$`)

type Bot struct {
	ID         int             `json:"id"`
	Type       BotType         `json:"type"`
	Query      string          `json:"query"`
	Blacklist  map[string]bool `json:"blacklist"`
	Headless   bool            `json:"headless"`
	PlayedAt   time.Time       `json:"played_at"`
	PlayResult string          `json:"play_result"`
	ChildID    int             `json:"child_id"`
}

func (b *Bot) MarshalZerologObject(evt *zerolog.Event) {
	evt.Int("#", b.ID).
		Str("q", b.Query).
		Any("t", b.Type).
		Any("x", b.Blacklist).
		Time("@", b.PlayedAt)
}

type BotType string

const (
	NewsBot    BotType = "news"
	SearchBot          = "search"
	SitemapBot         = "sitemap"
)

func (t *BotType) UnmarshalJSON(payload []byte) error {
	if text := strings.ReplaceAll(string(payload), `"`, ""); botTypeRegex.MatchString(text) {
		*t = BotType(text)
		return nil
	}
	return fmt.Errorf("unknown bot type: %s", payload)
}

type Blacklist map[string]bool

func (b *Blacklist) UnmarshalJSON(payload []byte) error {
	m := make(map[string]bool)
	text := strings.ReplaceAll(string(payload), `"`, "")
	for s := range strings.SplitSeq(text, `\n`) {
		if s = strings.TrimSpace(s); s != `null` && s != "" {
			m[s] = true
		}
	}
	*b = m
	return nil
}
