package bot

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/rs/zerolog"
)

var typeRegex = regexp.MustCompile(`^(news|search|sitemap)$`)

type Model struct {
	ID int `db:"id"`

	ChildID int `db:"child_id"`

	Type Type `db:"type"`

	Blacklist Blacklist `db:"blacklist"`

	Headless bool `db:"headless"`

	Query string `db:"query"`

	LastRunAt *time.Time `db:"last_run_at"`
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

func (t *Type) Scan(value any) error {
	if text := fmt.Sprintf("%v", value); typeRegex.MatchString(text) {
		*t = Type(text)
		return nil
	}
	return fmt.Errorf("unknown bot type: %v", value)
}

type Blacklist model.Set[string]

func (b *Blacklist) Scan(value any) error {
	ms := model.MakeSet[string]()
	if value != nil {
		for s := range strings.SplitSeq(fmt.Sprintf("%v", value), `\n`) {
			if s = strings.TrimSpace(s); s != `null` && s != "" {
				ms.Add(s)
			}
		}
	}
	*b = Blacklist(ms)
	return nil
}

func (b *Blacklist) OK(words []string) bool {
	s := model.Set[string](*b)
	return !slices.ContainsFunc(words, s.Has)
}
