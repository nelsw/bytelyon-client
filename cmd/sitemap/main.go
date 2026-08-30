package main

import (
	"bytelyon-client/internal/model"
	"bytelyon-client/internal/provider/logs"
	"bytelyon-client/internal/provider/play"
	"flag"
	"fmt"

	"github.com/mxschmitt/playwright-go"
)

func main() {

	logs.Init("trace")

	var domain string
	var depth, parallel int
	flag.StringVar(&domain, "domain", "bytelyon.com", "domain to crawl")
	flag.IntVar(&depth, "depth", 5, "links deep to follow")
	flag.IntVar(&parallel, "parallel", 25, "pages to crawl in parallel")
	flag.Parse()

	s := model.NewSitemap(domain,
		model.WithID(999),
		model.WithDepth(depth),
		model.WithParallelism(parallel),
	)
	play.It(false, func(ctx playwright.BrowserContext) {
		s.Build(ctx)
	})
	for _, k := range s.Keys() {
		fmt.Println(k)
	}
}
