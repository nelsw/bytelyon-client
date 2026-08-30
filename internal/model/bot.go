package model

import (
	"bytelyon-client/internal/provider/api"
	"encoding/json/v2"
	"regexp"
	"time"

	"github.com/rs/zerolog"
)

var botTypeRegex = regexp.MustCompile(`^(news|search|sitemap)$`)

func Bots() (arr []*Bot) {
	_ = json.Unmarshal(api.Get("bots"), &arr)
	return
}

type Bot struct {
	ID         int       `json:"id"`
	Type       BotType   `json:"type"`
	Query      string    `json:"query"`
	Blacklist  Blacklist `json:"blacklist"`
	Headless   bool      `json:"headless"`
	PlayedAt   time.Time `json:"played_at"`
	PlayResult string    `json:"play_result"`
	ChildID    int       `json:"child_id"`
}

func (b *Bot) MarshalZerologObject(evt *zerolog.Event) {
	evt.Int("#", b.ID).
		Str("q", b.Query).
		Any("t", b.Type).
		Any("x", b.Blacklist).
		Time("@", b.PlayedAt)
}

func (b *Bot) Save(result ...string) {
	b.PlayedAt = time.Now().UTC()
	if len(result) == 0 {
		b.PlayResult = "ok"
	} else {
		b.PlayResult = result[0]
	}
	api.Put(b, "bots", b.ID)
}
