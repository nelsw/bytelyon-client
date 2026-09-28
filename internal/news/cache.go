package news

import (
	"fmt"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/cache"
)

func jobCountKey(botID int) string {
	return fmt.Sprintf("%s:%d", bot.NewsType, botID)
}

func incr(botID int) {
	_ = cache.Incr(jobCountKey(botID))
}

func decr(botID int) {
	if i, _ := cache.Decr(jobCountKey(botID)); i == 0 {
		bot.Update(botID)
	}
}

func remove(botID int) {
	_ = cache.Del(jobCountKey(botID))
}
