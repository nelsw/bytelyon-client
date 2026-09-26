package main

import (
	"github.com/nelsw/bytelyon-client/internal/app"
	"github.com/nelsw/bytelyon-client/internal/bot"
)

func main() {
	app.Init(true)
	defer app.Close()
	app.HandleBot(bot.FindOne())
}
