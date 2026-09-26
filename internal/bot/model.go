package bot

import (
	"database/sql/driver"
	"fmt"
	"regexp"
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

func (t *Type) Value() (driver.Value, error) {
	return t.String(), nil
}

func (t *Type) String() string {
	return string(*t)
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
	for _, w := range words {
		if s.Has(w) {
			return false
		}
	}
	return true
}

func (m *Model) RanBefore(t time.Time) bool {
	return m.LastRunAt == nil || m.LastRunAt.Before(t)
}
