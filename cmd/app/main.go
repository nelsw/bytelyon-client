package main

import (
	"bytelyon-client/internal/config"
	"bytelyon-client/internal/model"
	"fmt"
	"maps"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

func main() {

	config.FromCLI()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	todo := make(map[int]*model.Bot)

	poller := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-quit:
			fmt.Println() // newline after Ctrl+C
			log.Log().Msg("quitting...")
			poller.Stop()
			log.Log().Msg("👋")
			return
		case <-poller.C:
			log.Log().Msg("polling...")
			for _, bot := range model.Bots() {
				if _, ok := todo[bot.ID]; !ok {
					todo[bot.ID] = bot
				}
			}
		default:
			for key := range maps.Keys(todo) {
				todo[key].Do()
				delete(todo, key)
			}
		}
	}
}
