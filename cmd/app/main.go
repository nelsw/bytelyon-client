package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/nelsw/bytelyon-client/pkg/logs"
)

func main() {
	logs.Init()
	//godump.Dump(bots.Find())
	//bots, err := bot.Query(bot.NewsType, bot.SitemapType)
	//if err != nil {
	//	panic(err)
	//}
	//godump.DumpJSON(bots)

	//j := news.New(&bot.Entity{
	//	ID:        rand.Intn(100),
	//	Type:      bot.NewsType,
	//	Blacklist: make(map[string]bool),
	//	Headless:  true,
	//	Query:     "lindsay clancy",
	//	ChildID:   rand.Intn(10),
	//	LastRunAt: pgtype.Timestamptz{},
	//})
	//j.Work()
	//godump.DumpJSON(j.Done)
}
