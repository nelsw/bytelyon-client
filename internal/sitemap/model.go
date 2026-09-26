package sitemap

import (
	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/model"
)

const maxDepth = 5
const maxAsync = 20

type Model struct {
	Bot                    *bot.Model
	ID                     int `json:"id"`
	*model.SyncSet[string] `json:"urls"`
}

func From(b *bot.Model) Model {
	return Model{
		Bot:     b,
		ID:      b.ChildID,
		SyncSet: model.NewSyncSet[string](),
	}
}
