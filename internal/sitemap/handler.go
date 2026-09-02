package sitemap

import (
	"github.com/nelsw/bytelyon-client/pkg/store"
	"github.com/nelsw/bytelyon-client/pkg/uuid"
)

func Handle(id, botID int, headless bool, domain string) {

	r := NewResult(id, domain)
	err := Fetch(r, headless)
	if err == nil {
		_ = store.Save(r.Keys(), "sitemap", botID, uuid.FromURL(r.URL).String()+".json")
	}
}
