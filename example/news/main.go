package main

import (
	"math/rand"
	"time"

	"github.com/nelsw/bytelyon-client/config"
	"github.com/nelsw/bytelyon-client/internal/news"
)

func main() {
	config.RunDry(true)
	news.Handle(
		rand.Intn(100),
		false,
		"btc forecast",
		time.Now().Add(-time.Hour*24),
		nil,
	)
}
