package sitemap

import (
	"github.com/nelsw/bytelyon-client/pkg/store"
	"github.com/nelsw/bytelyon-client/pkg/uuid"
)

func Handle(id, botID int, headless bool, domain string) error {

	r := NewResult(id, domain)
	err := Fetch(r, headless)
	if err != nil {
		panic(err)
	}
	urls := r.Keys()
	return store.Save(urls, "sitemap", botID, uuid.FromURL(r.URL).String()+".json")
}
