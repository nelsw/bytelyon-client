package main

import (
	"math/rand"
	"time"

	"github.com/nelsw/bytelyon-client/internal/news"
	"github.com/nelsw/bytelyon-client/pkg/model"
)

func main() {

	id := rand.Intn(100)
	query := "btc forecast"
	after := time.Now().Add(-time.Hour * 24 * 7 * 52)
	ignore := model.NewSet[string]()
	ignore.Put("test", true)

	err := news.Handle(id, false, query, after, ignore)
	if err != nil {
		panic(err)
	}
}
