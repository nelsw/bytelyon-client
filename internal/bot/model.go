package bot

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nelsw/bytelyon-client/pkg/model"
)

var typeRegex = regexp.MustCompile(`^(news|search|sitemap)$`)

type Model struct {
	ID int `db:"id"`

	ChildID int `db:"child_id"`

	Type `db:"type"`

	Blacklist `db:"blacklist"`

	Headless bool `db:"headless"`

	Query string `db:"query"`

	LastRunAt *time.Time `db:"last_run_at"`
}

//func (j *Job) MarshalZerologObject(evt *zerolog.Event) {
//	evt.Int("#", j.ID).
//		Str("q", j.Query).
//		Any("t", j.Type).
//		Any("x", j.Blacklist).
//		Time("@", j.WorkedAt)
//}

type Type string

const (
	NewsType    Type = "news"
	SearchType       = "search"
	SitemapType      = "sitemap"
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
