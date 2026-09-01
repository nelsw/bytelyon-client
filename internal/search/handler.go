package search

import (
	"github.com/nelsw/bytelyon-client/pkg/store"
	"github.com/nelsw/bytelyon-client/pkg/uuid"
)

func Handle(id, botID int, headless bool, query string, ignore []string) error {

	r := NewResult(id, botID, query, ignore)

	if err := Fetch(r, headless); err != nil {
		return err
	}

	return store.Save(r, "search", botID, uuid.FromURL(r.url).String()+".json")
}
