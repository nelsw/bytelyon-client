package bot

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var typeRegex = regexp.MustCompile(`^(news|search|sitemap)$`)

type Model struct {
	ID int `db:"id" json:"id"`

	ChildID int `db:"child_id" json:"child_id"`

	Type Type `db:"type" json:"type"`

	Blacklist Blacklist `db:"blacklist" json:"blacklist"`

	Headless bool `db:"headless" json:"headless"`

	Query string `db:"query" json:"query"`

	LastRunAt *time.Time `db:"last_run_at" json:"last_run_at"`
}

func (m *Model) LastRun() time.Time {
	if m.LastRunAt == nil {
		return time.Time{}
	}
	return *m.LastRunAt
}

func (m *Model) MarshalZerologObject(evt *zerolog.Event) {
	ranAt := "Never"
	if m.LastRunAt != nil {
		ranAt = m.LastRunAt.Format(time.DateTime)
	}
	evt.Int("#", m.ID).
		Str("q", m.Query).
		Any("t", m.Type).
		Any("x", m.Blacklist).
		Str("@", ranAt)
}

type Type string

const (
	NewsType    Type = "news"
	SearchType  Type = "search"
	SitemapType Type = "sitemap"
)

func (t *Type) String() string {
	return string(*t)
}

func (t *Type) Scan(value any) error {
	if text := fmt.Sprintf("%v", value); typeRegex.MatchString(text) {
		*t = Type(text)
		return nil
	}
	return fmt.Errorf("unknown bot type: %v", value)
}

type Blacklist map[string]bool

func (b *Blacklist) Scan(value any) error {
	*b = make(map[string]bool)
	if value == nil {
		return nil
	}
	for s := range strings.SplitSeq(fmt.Sprintf("%v", value), `\n`) {
		if s = strings.TrimSpace(s); s != `null` && s != "" {
			(*b)[s] = true
		}
	}
	return nil
}

func (b *Blacklist) OK(words []string) bool {
	return !slices.ContainsFunc(words, func(s string) bool {
		_, exists := (*b)[s]
		return exists
	})
}
