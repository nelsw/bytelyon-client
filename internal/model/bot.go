package model

import (
	"bytelyon-client/internal/provider/api"
	"encoding/json/v2"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Bots []*Bot

func GetBots() (arr Bots) {
	_ = json.Unmarshal(api.Get("bots"), &arr)
	return
}

type Bot struct {
	ID        int       `json:"id"`
	Type      BotType   `json:"type"`
	Query     string    `json:"query"`
	Blacklist Blacklist `json:"blacklist"`
	Headless  bool      `json:"headless"`
	LastRunAt time.Time `json:"last_run_at"`
	SitemapID int       `json:"sitemap_id,omitempty"`
	SearchID  int       `json:"serp_id,omitempty"`
}

func (b *Bot) MarshalZerologObject(evt *zerolog.Event) {
	evt.Int("#", b.ID).Str("q", b.Query).Any("t", b.Type)
}

func (b *Bot) Save() { api.Put(b, "bots", b.ID) }

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
	return "unkown"
}

func (t *BotType) UnmarshalJSON(payload []byte) error {
	if text := string(payload); text == `"news"` || text == `"search"` || text == `"sitemap"` {
		*t = BotType(strings.ReplaceAll(text, `"`, ""))
		return nil
	}
	return fmt.Errorf("unknown bot type: %s", payload)
}

type Blacklist map[string]bool

func (b *Blacklist) UnmarshalJSON(payload []byte) error {
	var arr []string
	if err := json.Unmarshal(payload, &arr); err != nil {
		return err
	}

	m := make(map[string]bool)
	for _, item := range arr {
		m[item] = true
	}
	*b = m
	return nil
}

func (b *Blacklist) OK(args ...string) bool {
	for _, arg := range args {
		if slices.ContainsFunc(strings.Split(arg, " "), func(s string) bool { return (*b)[s] }) {
			return false
		}
	}
	return true
}
